package state

import "time"

type State struct {
	ID                      string
	ResourceName            string
	ResourceID              string
	Status                  string
	StartTime               time.Time
	EndTime                 time.Time
	ExecutionDuration       time.Duration
	ExecutionDurationMillis int64
	ExecutionError          string
}
