package boltzmann

import "time"

// RepositoryConfig a basic and general-purposed repository configuration.
// Extend from this structure (with struct-embedding) if more fields are required for specific scenarios.
type RepositoryConfig struct {
	ItemTTL time.Duration // Time for an item to be preserved in the store (aka. Time-To-Live).
}
