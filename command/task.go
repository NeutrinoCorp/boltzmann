package command

type ScheduleTaskCommand struct {
	Driver         string            `json:"driver"`
	Resource       string            `json:"resource"`
	AgentArguments map[string]string `json:"agent_arguments"`
	Payload        []byte            `json:"payload"`
}
