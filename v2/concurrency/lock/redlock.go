package lock

import (
	"context"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
)

type Redlock struct {
	RedsyncClient *redsync.Redsync
	Config        DistributedLockConfig

	redsyncSingletonMu sync.Mutex
	mu                 *redsync.Mutex
}

var _ DistributedLock = &Redlock{}

func adaptRedsyncOptions(cfg DistributedLockConfig) []redsync.Option {
	opts := make([]redsync.Option, 0)
	if cfg.LeaseExpireDuration > time.Duration(0) {
		opts = append(opts, redsync.WithExpiry(cfg.LeaseExpireDuration))
	}
	return opts
}

func (r *Redlock) allocateLock() {
	r.redsyncSingletonMu.Lock()
	if r.mu == nil {
		r.mu = r.RedsyncClient.NewMutex(r.Config.Name, adaptRedsyncOptions(r.Config)...)
		r.redsyncSingletonMu.Unlock()
		return
	}
	r.redsyncSingletonMu.Unlock()
}

func (r *Redlock) Obtain(ctx context.Context) error {
	r.allocateLock()
	return r.mu.LockContext(ctx)
}

func (r *Redlock) Release(ctx context.Context) error {
	r.allocateLock()
	_, err := r.mu.UnlockContext(ctx)
	return err
}

func (r *Redlock) Extend(ctx context.Context) error {
	r.allocateLock()
	_, err := r.mu.ExtendContext(ctx)
	return err
}

type RedlockFactory struct {
	RedsyncClient *redsync.Redsync
	Config        DistributedLockConfig
}

var _ Factory = RedlockFactory{}

func (r RedlockFactory) NewLock(name string) (Lock, error) {
	r.Config.Name = name
	return &Redlock{
		RedsyncClient:      r.RedsyncClient,
		Config:             r.Config,
		redsyncSingletonMu: sync.Mutex{},
		mu:                 nil,
	}, nil
}
