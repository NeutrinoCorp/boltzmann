package task

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
)

type Service struct {
	AgentRegistry agent.Registry
}

func (s Service) RunTask(ctx context.Context, task boltzmann.Task) error {
	ag, err := s.AgentRegistry.Get(task.Driver)
	if err != nil {
		return err
	}

	return ag.ExecTask(ctx, task)
}
