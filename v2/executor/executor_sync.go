package executor

import (
	"context"

	"github.com/neutrinocorp/boltzmann/v2/executor/delegate"
)

type SyncExecutor[T any] struct {
	Delegate delegate.Delegate[T]
}

var _ Executor[string] = SyncExecutor[string]{}

func (s SyncExecutor[T]) ExecuteAll(ctx context.Context, tasks []T) error {
	for _, task := range tasks {
		if err := s.Delegate.Execute(ctx, task); err != nil {
			return err
		}
	}
	return nil
}
