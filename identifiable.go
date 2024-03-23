package boltzmann

// Identifiable adheres identification capabilities to a Go structure.
type Identifiable[T comparable] interface {
	// GetID generates or retrieves (depending on the concrete implementation) the structure's unique identifier.
	GetID() T
}

// NoopIdentifiable is the no-operation Identifiable struct.
type NoopIdentifiable struct {
}

var _ Identifiable[string] = NoopIdentifiable{}

func (n NoopIdentifiable) GetID() string {
	return ""
}
