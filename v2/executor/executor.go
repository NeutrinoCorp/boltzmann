package executor

import "context"

type Executor[T any] interface {
	ExecuteAll(ctx context.Context, tasks []T) error
}
