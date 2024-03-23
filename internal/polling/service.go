package polling

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/internal/executor"
	"github.com/neutrinocorp/boltzmann/internal/queue"
)

type Config struct {
	Name             string
	PollInterval     time.Duration
	RetryInterval    time.Duration
	MaxRetries       int
	BatchSizePerPoll int
}

type Service[T any] struct {
	Config          Config
	Queue           queue.Queue[T]
	ExecutorService executor.Executor[T]

	logger            zerolog.Logger
	baseCtx           context.Context
	baseCtxCancelFunc context.CancelFunc
}

var _ boltzmann.BackgroundProcess = &Service[string]{}

func (p *Service[T]) Start(ctx context.Context) error {
	p.baseCtx, p.baseCtxCancelFunc = context.WithCancel(ctx)
	p.logger = log.With().Str("poller", p.Config.Name).Logger()

	errCount := 0
mainLoop:
	for {
		select {
		case <-p.baseCtx.Done():
			p.logger.Info().Msg("stopping polling")
			break mainLoop
		default:
		}

		// TODO: Add message pop (actual remove). Popping directly might incur in data loss.
		//  Redis: Uses LRANGE, requires LTRIM to remove items from queue
		//  SQS: Uses GetMessages and then DeleteMessage APIs
		// Or: Supervisor will collaborate with Sched to retry. State storage will retain data, thus, no data loss.
		// State is committed before actual task scheduling (before writing data into queue).
		// SQS in Java has the behavior in listener annotations to delete message automatically once received
		// not mattering if the listener routine fails.
		p.logger.Info().Msg("fetching data from polling")
		tasks, err := p.Queue.Pop(ctx, p.Config.BatchSizePerPoll)
		if err != nil {
			errCount++
			p.logger.Err(err).Int("error_count", errCount).Msg("got error from polling")
			if errCount == p.Config.MaxRetries {
				return err
			}
			time.Sleep(p.Config.RetryInterval)
			continue
		} else if len(tasks) == 0 {
			time.Sleep(p.Config.PollInterval)
			continue
		}

		if err = p.ExecutorService.ExecuteAll(p.baseCtx, tasks); err != nil {
			errCount++
			p.logger.Err(err).Int("error_count", errCount).Msg("got error from polling")
			if errCount == p.Config.MaxRetries {
				return err
			}
			time.Sleep(p.Config.RetryInterval)
			continue
		}
		time.Sleep(p.Config.PollInterval)
	}
	return nil
}

func (p *Service[T]) Shutdown(_ context.Context) error {
	// TODO: Add graceful shutdown
	p.baseCtxCancelFunc()
	return nil
}
