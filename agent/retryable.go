package agent

import (
	"context"
	"io"

	"github.com/cenkalti/backoff/v4"

	"github.com/neutrinocorp/boltzmann"
)

// Retryable is an Agent proxy used by other agents to attach retry features without requiring further
// modification.
type Retryable struct {
	Next Agent
}

var _ Middleware = &Retryable{}

func (r *Retryable) Execute(ctx context.Context, task boltzmann.Task) (io.ReadCloser, error) {
	var res io.ReadCloser
	err := backoff.Retry(func() error {
		resExec, errExec := r.Next.Execute(ctx, task)
		if errExec != nil {
			return errExec
		}
		res = resExec
		return nil
	}, backoff.NewExponentialBackOff())
	return res, err
}

func (r *Retryable) SetNext(a Agent) {
	r.Next = a
}
