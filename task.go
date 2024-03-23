package boltzmann

import (
	"strconv"

	"github.com/neutrinocorp/boltzmann/internal/id"
)

// Task a unit of work to be executed by a Boltzmann agent.
type Task struct {
	TaskID          int               // Task unique identifier.
	ExecutionPlanID string            // Correlation StateID.
	Driver          string            // Driver type to be used by agent executors.
	ResourceURL     string            // Location of the resource as URL.
	AgentArguments  map[string]string // Arguments to be passed to the execution agent.
	TypeMIME        string            // MIME Type of the payload.
	Payload         any               `json:"-"` // Transient data to be passed to internal components for different utilities.
	EncodedPayload  []byte            // Encoded data to be passed to the execution agent.
}

var _ Identifiable[string] = Task{}

func (t Task) GetID() string {
	return id.NewCompositeKey(t.ExecutionPlanID, strconv.Itoa(t.TaskID))
}
