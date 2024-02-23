package lock

import "context"

type Lock interface {
	Obtain(ctx context.Context) error
	Release(ctx context.Context) error
}
