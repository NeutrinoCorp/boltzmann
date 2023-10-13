package queue

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/state"
)

var stateUpdLogger = log.With().Str("component", "queue.middlewares.state_updater").Logger()

type StateUpdaterMiddleware struct {
	Repository state.Repository
	Next       Queue
}

var _ Queue = StateUpdaterMiddleware{}

func (s StateUpdaterMiddleware) Push(ctx context.Context, tasks ...boltzmann.Task) error {
	if err := s.Next.Push(ctx, tasks...); err != nil {
		return err
	}

	go func() {
		scopedCtx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		if err := s.Repository.SaveAll(scopedCtx, tasks...); err != nil {
			stateUpdLogger.Err(err).Msg("failed to save task states")
		}
	}()
	return nil
}

func (s StateUpdaterMiddleware) Pop(ctx context.Context) ([]boltzmann.Task, error) {
	return s.Next.Pop(ctx)
}
