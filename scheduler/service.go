package scheduler

import (
	"context"
	"time"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/command"
	"github.com/neutrinocorp/boltzmann/state"

	"github.com/segmentio/ksuid"
)

type Service struct {
	Scheduler       TaskScheduler
	StateRepository state.Repository
}

func (s Service) Schedule(ctx context.Context, commands []command.ScheduleTaskCommand) []ScheduleTaskResult {
	correlationID := ksuid.New().String()
	tasks := make([]boltzmann.Task, 0, len(commands))
	for _, cmd := range commands {
		taskID := ksuid.New().String()
		tasks = append(tasks, boltzmann.Task{
			TaskID:           taskID,
			CorrelationID:    correlationID,
			Driver:           cmd.Driver,
			ResourceLocation: cmd.Resource,
			AgentArguments:   cmd.AgentArguments,
			Payload:          cmd.Payload,
			Status:           boltzmann.TaskStatusInit,
			StartTime:        time.Now().UTC(),
		})
	}
	return s.Scheduler.Schedule(ctx, tasks)
}

func (s Service) GetTaskState(ctx context.Context, taskID string) (boltzmann.Task, error) {
	return s.StateRepository.Get(ctx, taskID)
}
