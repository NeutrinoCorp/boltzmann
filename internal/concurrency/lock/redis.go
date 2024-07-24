package lock

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/modern-go/reflect2"
)

// RedisLock is the Redis implementation of DistributedLock using the `redlock` algorithm.
//
// More information about redlock algorithm here: https://redis.io/docs/manual/patterns/distributed-locks/.
type RedisLock struct {
	Mu *redsync.Mutex
}

var _ DistributedLock = &RedisLock{}

func NewRedisLock[T any](cfg DistributedLockConfig, rds *redsync.Redsync) RedisLock {
	var zeroVal T
	typeOfStr := reflect2.TypeOf(zeroVal).String()
	return RedisLock{
		Mu: rds.NewMutex(typeOfStr, adaptRedsyncOptions(cfg)...),
	}
}

func adaptRedsyncOptions(cfg DistributedLockConfig) []redsync.Option {
	opts := make([]redsync.Option, 0)
	if cfg.LeaseDuration > time.Duration(0) {
		opts = append(opts, redsync.WithExpiry(cfg.LeaseDuration))
	}
	return opts
}

func (r RedisLock) Obtain(ctx context.Context) error {
	return r.Mu.LockContext(ctx)
}

func (r RedisLock) Release(ctx context.Context) error {
	_, err := r.Mu.UnlockContext(ctx)
	return err
}

func (r RedisLock) Extend(ctx context.Context) error {
	_, err := r.Mu.ExtendContext(ctx)
	return err
}
