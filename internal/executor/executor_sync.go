package executor

import (
	"context"

	"github.com/neutrinocorp/boltzmann/internal/executor/delegate"
)

// SyncExecutor is the synchronous implementation of Executor.
// It executes tasks in sequence and no extra goroutines are used.
type SyncExecutor[T any] struct {
	Delegate delegate.Delegate[T]
}

var _ Executor[string] = SyncExecutor[string]{}

func (s SyncExecutor[T]) ExecuteAll(ctx context.Context, args []T) error {
	for _, arg := range args {
		if err := s.Delegate.Execute(ctx, arg); err != nil {
			return err
		}
	}
	return nil
}
