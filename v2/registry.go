package boltzmann

import (
	"fmt"
)

type ErrRegistryEntryAlreadyExists struct {
	RegistryName string
	Key          string
}

var _ error = ErrRegistryEntryAlreadyExists{}

func (e ErrRegistryEntryAlreadyExists) Error() string {
	return fmt.Sprintf("registry.%s: entry <<%s>> already exists", e.RegistryName, e.Key)
}

type ErrRegistryEntryNotFound struct {
	RegistryName string
	Key          string
}

var _ error = ErrRegistryEntryNotFound{}

func (e ErrRegistryEntryNotFound) Error() string {
	return fmt.Sprintf("registry.%s: entry <<%s>> not found", e.RegistryName, e.Key)
}

type Registry[T any] interface {
	Register(key string, component T) error
	Get(key string) (T, error)
	Exists(key string) error
}
