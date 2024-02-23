package boltzmann

import "fmt"

// Task a unit of work to be executed by a Boltzmann agent.
type Task struct {
	TaskID          int
	ExecutionPlanID string            // Correlation ID.
	Driver          string            // Driver type to be used by agent executors.
	ResourceURL     string            // Location of the resource as URL.
	AgentArguments  map[string]string // Arguments to be passed to the execution agent.
	TypeMIME        string            // MIME Type of the payload.
	Payload         any               `json:"-"` // Data to be passed to be later encoded.
	EncodedPayload  []byte            // Encoded data to be passed to the execution agent.
}

var _ Identifiable[string] = Task{}

func (t Task) GetID() string {
	return fmt.Sprintf("%s&%d", t.ExecutionPlanID, t.TaskID)
}
