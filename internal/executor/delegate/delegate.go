package delegate

import "context"

// Delegate is a function/routine interface executed by executor.Executor(s) instances.
type Delegate[T any] interface {
	// Execute executes the underlying task(s).
	Execute(ctx context.Context, arg T) error
}
