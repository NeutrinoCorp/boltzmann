package agent

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

type Agent interface {
	Execute(ctx context.Context, task boltzmann.Task) error
}

type Middleware interface {
	Agent
	SetNext(a Agent)
}
