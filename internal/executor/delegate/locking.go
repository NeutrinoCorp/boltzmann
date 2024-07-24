package delegate

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/internal/concurrency/lock"
)

// LockingMiddleware is an implementation of Delegate
// that performs concurrency-locking mechanisms for each process execution.
type LockingMiddleware[I comparable, T boltzmann.Identifiable[I]] struct {
	Lock lock.Lock
	Next Delegate[T]
}

var _ Delegate[boltzmann.NoopIdentifiable] = LockingMiddleware[string, boltzmann.NoopIdentifiable]{}

func (d LockingMiddleware[I, T]) Execute(ctx context.Context, arg T) (err error) {
	// TODO: Add checksum instead id
	resourceIDStr := fmt.Sprintf("%v", arg.GetID())
	lockKey := fmt.Sprintf("lock#%s", resourceIDStr)

	log.Info().Str("resource_id", resourceIDStr).Str("lock_key", lockKey).Msg("acquiring lock")
	if err = d.Lock.Obtain(ctx); err != nil {
		return err
	}
	defer func() {
		log.Info().Str("resource_id", resourceIDStr).Str("lock_key", lockKey).Msg("releasing lock")
		if errLock := d.Lock.Release(ctx); errLock != nil {
			log.Err(errLock).Str("resource_id", resourceIDStr).
				Str("lock_key", lockKey).Msg("cannot release lock")
		}
	}()

	err = d.Next.Execute(ctx, arg)
	return
}
