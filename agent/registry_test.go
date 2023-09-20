package agent_test

import (
	"testing"

	"github.com/neutrinocorp/boltzmann/agent"
)

func TestRegistry_AddMiddleware(t *testing.T) {
	reg := agent.NewRegistry()
	reg.AddMiddleware(&middlewareFake{})
	reg.AddMiddleware(&middlewareFake{})
	reg.AddMiddleware(&middlewareFake{})
	reg.Register("http", agent.HTTP{})
}
