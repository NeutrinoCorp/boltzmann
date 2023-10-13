package queue

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

type Queue interface {
	// Push appends a set of tasks into the queue. If FIFO, task will be appended at the tail. If LIFO, task will be appended
	// at the head.
	Push(ctx context.Context, task ...boltzmann.Task) error
	// Pop retrieves several tasks from the queue.
	Pop(ctx context.Context) ([]boltzmann.Task, error)
}
