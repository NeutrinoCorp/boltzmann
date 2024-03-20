package controller

type TaskResponse struct {
	TaskID          int               `json:"task_id"`
	ExecutionPlanID string            `json:"execution_plan_id"`
	Driver          string            `json:"driver"`          // Driver type to be used by agent executors.
	ResourceURL     string            `json:"resource_url"`    // Location of the resource as URL.
	AgentArguments  map[string]string `json:"agent_arguments"` // Arguments to be passed to the execution agent.
	TypeMIME        string            `json:"mime_type"`       // MIME Type of the payload.
	PayloadSize     int               `json:"payload_size"`
}

type ExecutionPlanResponse struct {
	PlanID       string         `json:"plan_id"`
	WithFairness bool           `json:"with_fairness"`
	Tasks        []TaskResponse `json:"tasks"`
}
