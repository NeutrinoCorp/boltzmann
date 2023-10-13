package agent_test

import (
	"context"
	"io"
	"testing"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
)

type middlewareFake struct {
	t    *testing.T
	next agent.Agent
}

var _ agent.Middleware = &middlewareFake{}

func (m *middlewareFake) Execute(_ context.Context, _ boltzmann.Task) (io.ReadCloser, error) {
	m.t.Log("executing fake middleware")
	return nil, nil
}

func (m *middlewareFake) SetNext(a agent.Agent) {
	m.next = a
}
