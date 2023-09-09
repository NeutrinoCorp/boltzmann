package boltzmann

import (
	"encoding"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Task struct {
	TaskID            string
	CorrelationID     string
	Driver            string
	ResourceURI       string
	AgentArguments    map[string]string
	Payload           []byte
	Status            TaskStatus
	SuccessMessage    string
	FailureMessage    string
	StartTime         time.Time
	EndTime           time.Time
	ExecutionDuration time.Duration
}

var _ encoding.BinaryMarshaler = Task{}

var _ encoding.BinaryUnmarshaler = Task{}

func (t Task) MarshalBinary() (data []byte, err error) {
	return jsoniter.Marshal(t)
}

func (t Task) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, &t)
}
