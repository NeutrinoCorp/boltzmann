package delegate

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/concurrency/lock"
)

type LockingMiddleware[I comparable, T boltzmann.Identifiable[I]] struct {
	LockFactory lock.Factory
	Next        Delegate[T]
}

var _ Delegate[boltzmann.ExecutionPlan] = LockingMiddleware[string, boltzmann.ExecutionPlan]{}

func (d LockingMiddleware[I, T]) Execute(ctx context.Context, item T) (err error) {
	resourceIDStr := fmt.Sprintf("%v", item.GetID())
	lockKey := fmt.Sprintf("lock#%s", resourceIDStr)
	l, err := d.LockFactory.NewLock(lockKey)
	if err != nil {
		return err
	}

	log.Info().Str("resource_id", resourceIDStr).Str("lock_key", lockKey).Msg("acquiring lock")
	if err = l.Obtain(ctx); err != nil {
		return err
	}
	defer func() {
		log.Info().Str("resource_id", resourceIDStr).Str("lock_key", lockKey).Msg("releasing lock")
		if errLock := l.Release(ctx); errLock != nil {
			log.Err(errLock).Str("resource_id", resourceIDStr).
				Str("lock_key", lockKey).Msg("cannot release lock")
		}
	}()

	err = d.Next.Execute(ctx, item)
	return
}
