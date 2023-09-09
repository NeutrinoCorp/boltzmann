package queue

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

// A Service is a queueing service component used by Boltzmann to load balance and distribute tasks to several
// workers (Boltzmann nodes).
type Service interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Enqueue(ctx context.Context, task boltzmann.Task) error
}
