package agent

import (
	"context"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/executor/delegate"
)

type ExecutorDelegate struct {
	Registry Registry
}

var _ delegate.Delegate[boltzmann.Task] = ExecutorDelegate{}

func (e ExecutorDelegate) Execute(ctx context.Context, task boltzmann.Task) error {
	ag, err := e.Registry.Get(task.Driver)
	if err != nil {
		return err
	}

	return ag.ExecTask(ctx, task)
}
