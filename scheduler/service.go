package scheduler

import (
	"context"
	"time"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/command"
	"github.com/neutrinocorp/boltzmann/factory"
	"github.com/neutrinocorp/boltzmann/state"
)

type Service struct {
	Scheduler       TaskScheduler
	StateRepository state.Repository
	FactoryID       factory.Identifier
}

func (s Service) Schedule(ctx context.Context, commands []command.ScheduleTaskCommand) ([]ScheduleTaskResult, error) {
	correlationID, err := s.FactoryID.NewID()
	if err != nil {
		return nil, err
	}
	tasks := make([]boltzmann.Task, 0, len(commands))
	for _, cmd := range commands {
		taskID, errID := s.FactoryID.NewID()
		if errID != nil {
			return nil, errID
		}
		tasks = append(tasks, boltzmann.Task{
			TaskID:         taskID,
			CorrelationID:  correlationID,
			Driver:         cmd.Driver,
			ResourceURI:    cmd.ResourceURI,
			AgentArguments: cmd.AgentArguments,
			Payload:        cmd.Payload,
			Status:         boltzmann.TaskStatusScheduled,
			ScheduleTime:   time.Now().UTC(),
		})
	}
	return s.Scheduler.Schedule(ctx, tasks)
}

func (s Service) GetTaskState(ctx context.Context, taskID string) (boltzmann.Task, error) {
	return s.StateRepository.Get(ctx, taskID)
}
