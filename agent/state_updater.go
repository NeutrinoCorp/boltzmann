package agent

import (
	"context"
	"time"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/state"
)

type StateUpdater struct {
	StateRepository state.Repository
	next            Agent
}

var _ Middleware = &StateUpdater{}

func (s *StateUpdater) SetNext(a Agent) {
	s.next = a
}

func (s *StateUpdater) Execute(ctx context.Context, task boltzmann.Task) (err error) {
	defer func() {
		task.Status = boltzmann.TaskStatusSucceed
		if err != nil {
			task.Status = boltzmann.TaskStatusFailed
			task.FailureMessage = err.Error()
		}

		task.EndTime = time.Now().UTC()
		task.ExecutionDuration = task.EndTime.Sub(task.StartTime)
		if errCommit := s.StateRepository.Save(ctx, task); errCommit != nil {
			logger.Err(errCommit).
				Str("task_id", task.TaskID).
				Str("driver", task.Driver).
				Str("resource_location", task.ResourceURI).
				Msg("failed to save state")
		}
	}()

	task.Status = boltzmann.TaskStatusPending
	if errCommit := s.StateRepository.Save(ctx, task); errCommit != nil {
		logger.Err(errCommit).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to save state")
		return
	}

	err = s.next.Execute(ctx, task)
	return
}
