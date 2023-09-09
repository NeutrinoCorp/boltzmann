package command

type ScheduleTaskCommand struct {
	Driver         string            `json:"driver"`
	ResourceURI    string            `json:"resource_uri"`
	AgentArguments map[string]string `json:"agent_arguments"`
	Payload        []byte            `json:"payload"`
}
