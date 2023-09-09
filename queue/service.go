package queue

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

type Service interface {
	Enqueue(ctx context.Context, task boltzmann.Task) error
}
