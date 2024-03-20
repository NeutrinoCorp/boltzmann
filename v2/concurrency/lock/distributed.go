package lock

import (
	"context"
	"time"
)

type DistributedLockConfig struct {
	Name                string
	LeaseExpireDuration time.Duration
}

type DistributedLock interface {
	Lock
	Extend(ctx context.Context) error
}
