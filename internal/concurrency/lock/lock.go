package lock

import "context"

// Lock is a core component used to lock a process goroutine/thread until unlocked.
// This makes other routine calls to wait for the process to get the Lock unlocked.
type Lock interface {
	// Obtain obtains a lock. Will sleep if the lock was previously obtained and not released.
	Obtain(ctx context.Context) error
	// Release releases the lock so other goroutines can run the process as well.
	Release(ctx context.Context) error
}
