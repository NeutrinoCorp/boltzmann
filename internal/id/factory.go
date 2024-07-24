package id

// FactoryFunc generates a unique identifier.
type FactoryFunc func() (string, error)
