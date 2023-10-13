package boltzmann_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neutrinocorp/boltzmann"
)

func TestNewTaskStatus(t *testing.T) {
	tests := []struct {
		name string
		in   string
		exp  boltzmann.TaskStatus
	}{
		{
			name: "empty",
			in:   "",
			exp:  boltzmann.TaskStatus(0),
		},
		{
			name: "scheduled",
			in:   "SCHEDULED",
			exp:  boltzmann.TaskStatus(1),
		},
		{
			name: "started",
			in:   "STARTED",
			exp:  boltzmann.TaskStatus(2),
		},
		{
			name: "failed",
			in:   "FAILED",
			exp:  boltzmann.TaskStatus(3),
		},
		{
			name: "succeed",
			in:   "SUCCEED",
			exp:  boltzmann.TaskStatus(4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := boltzmann.NewTaskStatus(tt.in)
			assert.Equal(t, tt.exp, out)
		})
	}
}

func TestTaskStatus_GoString(t *testing.T) {
	tests := []struct {
		name string
		in   boltzmann.TaskStatus
		exp  string
	}{
		{
			name: "empty",
			in:   boltzmann.TaskStatus(0),
			exp:  "",
		},
		{
			name: "scheduled",
			in:   boltzmann.TaskStatus(1),
			exp:  "SCHEDULED",
		},
		{
			name: "started",
			in:   boltzmann.TaskStatus(2),
			exp:  "STARTED",
		},
		{
			name: "failed",
			in:   boltzmann.TaskStatus(3),
			exp:  "FAILED",
		},
		{
			name: "succeed",
			in:   boltzmann.TaskStatus(4),
			exp:  "SUCCEED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.exp, tt.in.GoString())
		})
	}
}
