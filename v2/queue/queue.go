package queue

import "context"

type Queue[T any] interface {
	Push(ctx context.Context, items ...T) error
	Pop(ctx context.Context, popRange int) ([]T, error)
}
