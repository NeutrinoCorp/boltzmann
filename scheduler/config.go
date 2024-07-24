package scheduler

// ServiceConfig scheduler service configuration parameters.
type ServiceConfig struct {
	MaxScheduledTasks int `validate:"min=1,max=2048"` // Maximum number of tasks to be schedule by the scheduler component.
	// PayloadTruncateLimit Truncation limit (in bytes) of agent payloads.
	// It will be ignored if the value is less or equal than 0
	PayloadTruncateLimit int64
}
