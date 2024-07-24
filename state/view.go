package state

import "time"

type View struct {
	StateID                 string        `json:"state_id"`
	ResourceName            string        `json:"resource_name"`
	ResourceID              string        `json:"resource_id"`
	Status                  string        `json:"status"`
	StartTime               time.Time     `json:"start_time"`
	StartTimeMillis         int64         `json:"start_time_millis"`
	EndTime                 time.Time     `json:"end_time"`
	EndTimeMillis           int64         `json:"end_time_millis"`
	ExecutionDuration       time.Duration `json:"execution_duration"`
	ExecutionDurationMillis int64         `json:"execution_duration_millis"`
	ExecutionError          string        `json:"execution_error"`
}
