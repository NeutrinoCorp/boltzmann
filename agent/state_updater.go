package agent

import (
	"context"
	"io"
	"time"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/config"
	"github.com/neutrinocorp/boltzmann/state"
)

type StateUpdaterConfig struct {
	ResponseTruncateLimit int64
}

func setStateUpdaterConfigDefault() {
	config.SetDefault(config.StateTruncateLimit, int64(1024))
}

func NewStateUpdaterConfig() StateUpdaterConfig {
	setStateUpdaterConfigDefault()
	return StateUpdaterConfig{
		ResponseTruncateLimit: config.Get[int64](config.StateTruncateLimit),
	}
}

type StateUpdater struct {
	StateRepository state.Repository
	Config          StateUpdaterConfig
	next            Agent
}

var _ Middleware = &StateUpdater{}

func (s *StateUpdater) SetNext(a Agent) {
	s.next = a
}

func (s *StateUpdater) Execute(ctx context.Context, task boltzmann.Task) (res io.ReadCloser, err error) {
	task.StartTime = time.Now().UTC()
	defer func() {
		task.Status = boltzmann.TaskStatusSucceed
		if err != nil {
			task.Status = boltzmann.TaskStatusFailed
			task.FailureMessage = err.Error()
		}

		task.EndTime = time.Now().UTC()
		task.ExecutionDuration = task.EndTime.Sub(task.StartTime)
		reader := io.LimitReader(res, s.Config.ResponseTruncateLimit)
		resBytes, errRead := io.ReadAll(reader)
		if errRead == nil {
			task.Response = resBytes
			_ = res.Close()
		}
		if errCommit := s.StateRepository.Save(ctx, task); errCommit != nil {
			logger.Err(errCommit).
				Str("task_id", task.TaskID).
				Str("driver", task.Driver).
				Str("resource_location", task.ResourceURI).
				Msg("failed to save state")
		}
	}()

	task.Status = boltzmann.TaskStatusStarted
	if errCommit := s.StateRepository.Save(ctx, task); errCommit != nil {
		logger.Err(errCommit).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to save state")
		return
	}

	res, err = s.next.Execute(ctx, task)
	return
}
