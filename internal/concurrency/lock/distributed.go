package lock

import (
	"context"
)

// DistributedLock is a type of Lock that communicates to other nodes to coordinate a specific lock.
//
// Most of these kinds of locks get released automatically to avoid distributed deadlocks.
type DistributedLock interface {
	Lock
	// Extend extends the Lock automatic expiration time.
	Extend(ctx context.Context) error
}
