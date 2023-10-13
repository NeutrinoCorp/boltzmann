package boltzmann

import "time"

type Task struct {
	TaskID            string
	CorrelationID     string
	Driver            string
	ResourceURI       string
	AgentArguments    map[string]string
	Payload           []byte
	Status            TaskStatus
	Response          []byte
	FailureMessage    string
	ScheduleTime      time.Time
	StartTime         time.Time
	EndTime           time.Time
	ExecutionDuration time.Duration
}
