package agent

import (
	"context"
	"io"

	"github.com/neutrinocorp/boltzmann"
)

type Agent interface {
	Execute(ctx context.Context, task boltzmann.Task) (io.ReadCloser, error)
}

type Middleware interface {
	Agent
	SetNext(a Agent)
}
