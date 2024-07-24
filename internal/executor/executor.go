package executor

import "context"

// Executor is a core component that allows routines to execute delegate.Delegate routines with specified arguments
// (T) in an abstract way.
// It will depend on which kind of concrete implementation how these delegate routines will be executed.
//
// Use ConcurrentExecutor to execute routines in a concurrent way.
// On the other hand, use SyncExecutor to execute tasks sequentially.
type Executor[T any] interface {
	// ExecuteAll executes all tasks.
	ExecuteAll(ctx context.Context, args []T) error
}
