package agent

import (
	"net/http"

	"github.com/neutrinocorp/boltzmann"
)

// TODO: Add configurer to attach agent instances available through Registry API.
const registryName = "agent"

var registryMap = map[string]Agent{
	"http": HTTP{Client: http.DefaultTransport},
	"noop": Noop{},
}

// Registry manages Agent instances of the system.
type Registry struct {
}

var _ boltzmann.Registry[Agent] = Registry{}

// Register saves a component with the given key.
func (r Registry) Register(key string, component Agent) error {
	_, ok := registryMap[key]
	if ok {
		return boltzmann.ErrItemAlreadyExists{
			ResourceName: registryName,
			ResourceKey:  key,
		}
	}

	registryMap[key] = component
	return nil
}

// Get retrieves a component using its key.
func (r Registry) Get(key string) (Agent, error) {
	ag, ok := registryMap[key]
	if !ok {
		return nil, boltzmann.ErrItemNotFound{
			ResourceName: registryName,
			ResourceKey:  key,
		}
	}

	return ag, nil
}

// Exists indicates whether a component with the given key exists.
func (r Registry) Exists(key string) error {
	_, ok := registryMap[key]
	if !ok {
		return boltzmann.ErrItemNotFound{
			ResourceName: registryName,
			ResourceKey:  key,
		}
	}
	return nil
}
