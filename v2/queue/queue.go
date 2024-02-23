package queue

import "context"

type Queue[T any] interface {
	Push(ctx context.Context, queueName string, items []T) error
}
