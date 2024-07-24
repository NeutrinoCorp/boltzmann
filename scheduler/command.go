package scheduler

// TaskCommand represents a task to be performed by the system.
type TaskCommand struct {
	Driver         string            `json:"driver" validate:"required"`                        // Driver type to be used by agent executors.
	ResourceURL    string            `json:"resource_url" validate:"url"`                       // Location of the resource as URL.
	AgentArguments map[string]string `json:"agent_arguments" validate:"omitempty,min=1,max=25"` // Arguments to be passed to the execution agent.
	TypeMIME       string            `json:"mime_type" validate:"required"`                     // MIME Type of the payload.
	Payload        any               `json:"payload"`                                           // Data to be passed to the execution agent.
}

// ScheduleTasksCommand a set of tasks to be scheduled to be later executed by the system.
type ScheduleTasksCommand struct {
	Tasks        []TaskCommand `json:"tasks" validate:"min=1"` // Set of tasks to be executed.
	WithFairness bool          `json:"with_fairness"`          // Indicates whether to respect ordering of the given set of tasks or not.
}
