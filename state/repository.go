package state

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

type Repository interface {
	Save(ctx context.Context, task boltzmann.Task) error
	SaveAll(ctx context.Context, tasks ...boltzmann.Task) error
	Get(ctx context.Context, taskId string) (boltzmann.Task, error)
}
