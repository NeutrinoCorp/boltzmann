package scheduler

// TaskCommand represents a task to be performed by the system.
type TaskCommand struct {
	Driver      string            `json:"driver"`       // Driver type to be used by agent executors.
	ResourceURL string            `json:"resource_url"` // Location of the resource as URL.
	Arguments   map[string]string `json:"arguments"`    // Arguments to be passed to the execution agent.
	Payload     []byte            `json:"payload"`      // Data to be passed to the execution agent.
}

// ScheduleTasksCommand a set of tasks to be scheduled to be later executed by the system.
type ScheduleTasksCommand struct {
	Tasks        []TaskCommand `json:"tasks"`         // Set of tasks to be executed.
	WithFairness bool          `json:"with_fairness"` // Indicates whether to respect ordering of the given set of tasks or not.
}
