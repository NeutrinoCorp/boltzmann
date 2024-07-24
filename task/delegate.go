package task

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/internal/executor/delegate"
)

type Delegate struct {
	Service Service
}

var _ delegate.Delegate[boltzmann.Task] = Delegate{}

func (e Delegate) Execute(ctx context.Context, task boltzmann.Task) error {
	return e.Service.RunTask(ctx, task)
}
