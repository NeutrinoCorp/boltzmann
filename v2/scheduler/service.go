package scheduler

import (
	"context"

	"github.com/neutrinocorp/boltzmann/v2/codec"
	"github.com/neutrinocorp/boltzmann/v2/queue"
)

type ServiceConfig struct {
	QueueName string
}

type Service struct {
	Queue         queue.Queue[TaskCommand]
	Config        ServiceConfig
	CodecStrategy codec.Strategy
}

func (s Service) Schedule(ctx context.Context, cmd ScheduleTasksCommand) error {
	if len(cmd.Tasks) == 0 {
		return ErrNoExecutionPlan
	}

	// TODO: Separate FIFO from unordered queue, create another interface
	return s.Queue.Push(ctx, s.Config.QueueName, cmd.Tasks)
}
