package agent

import (
	"net/http"

	"github.com/neutrinocorp/boltzmann/v2"
)

const registryName = "agent"

var registryMap = map[string]Agent{
	"http": HTTP{Client: http.DefaultTransport},
	"noop": Noop{},
}

type Registry struct {
}

var _ boltzmann.Registry[Agent] = Registry{}

func (r Registry) Register(key string, component Agent) error {
	_, ok := registryMap[key]
	if ok {
		return boltzmann.ErrRegistryEntryAlreadyExists{
			RegistryName: registryName,
			Key:          key,
		}
	}

	registryMap[key] = component
	return nil
}

func (r Registry) Get(key string) (Agent, error) {
	ag, ok := registryMap[key]
	if !ok {
		return nil, boltzmann.ErrRegistryEntryNotFound{
			RegistryName: registryName,
			Key:          key,
		}
	}

	return ag, nil
}

func (r Registry) Exists(key string) error {
	_, ok := registryMap[key]
	if !ok {
		return boltzmann.ErrRegistryEntryNotFound{
			RegistryName: registryName,
			Key:          key,
		}
	}
	return nil
}
