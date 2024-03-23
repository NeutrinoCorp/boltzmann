package lock

import "time"

// DistributedLockConfig configuration for DistributedLock components.
type DistributedLockConfig struct {
	LeaseDuration time.Duration // Duration of a lock to get released automatically.
}
