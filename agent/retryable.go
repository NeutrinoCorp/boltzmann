package agent

import (
	"context"

	"github.com/cenkalti/backoff/v4"

	"github.com/neutrinocorp/boltzmann"
)

// Retryable is an Agent proxy used by other agents to attach retry features without requiring further
// modification.
type Retryable struct {
	Next Agent
}

var _ Agent = Retryable{}

func (r Retryable) Execute(ctx context.Context, task boltzmann.Task) error {
	return backoff.Retry(func() error {
		return r.Next.Execute(ctx, task)
	}, backoff.NewExponentialBackOff())
}
