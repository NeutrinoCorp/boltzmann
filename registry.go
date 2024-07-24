package boltzmann

// Registry a generic registry to manage a specific set of resources of the system.
//
// NOTE: Concurrency safeness depends on concrete implementations of this interface.
// DO NOT consider every registry as concurrent-safe.
type Registry[T any] interface {
	// Register saves a component with the given key.
	Register(key string, component T) error
	// Get retrieves a component using its key.
	Get(key string) (T, error)
	// Exists indicates whether a component with the given key exists.
	Exists(key string) error
}
