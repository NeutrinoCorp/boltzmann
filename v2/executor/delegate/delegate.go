package delegate

import "context"

type Delegate[T any] interface {
	Execute(ctx context.Context, item T) error
}
