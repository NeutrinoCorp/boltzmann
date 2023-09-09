package boltzmann

import (
	"encoding"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type TaskStatus uint8

const (
	_ TaskStatus = iota
	TaskStatusInit
	TaskStatusScheduled
	TaskStatusPending
	TaskStatusFailed
	TaskStatusSucceed
)

var taskStatusMap = map[TaskStatus]string{
	TaskStatusInit:      "INITIATED",
	TaskStatusScheduled: "SCHEDULED",
	TaskStatusPending:   "PENDING",
	TaskStatusFailed:    "FAILED",
	TaskStatusSucceed:   "SUCCEED",
}

var taskStatusMapBackward = map[string]TaskStatus{
	"INITIATED": TaskStatusInit,
	"SCHEDULED": TaskStatusScheduled,
	"PENDING":   TaskStatusPending,
	"FAILED":    TaskStatusFailed,
	"SUCCEED":   TaskStatusFailed,
}

type Task struct {
	TaskID            string
	CorrelationID     string
	Driver            string
	ResourceLocation  string
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
