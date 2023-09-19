package marshal

import (
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/neutrinocorp/boltzmann"
)

func MarshalTaskRedisStream(task boltzmann.Task) map[string]string {
	argsJSON, _ := jsoniter.Marshal(task.AgentArguments)
	return map[string]string{
		"task_id":            task.TaskID,
		"correlation_id":     task.CorrelationID,
		"driver":             task.Driver,
		"resource_uri":       task.ResourceURI,
		"agent_arguments":    string(argsJSON),
		"payload":            string(task.Payload),
		"status":             task.Status.String(),
		"success_message":    task.SuccessMessage,
		"failure_message":    task.FailureMessage,
		"start_time":         task.StartTime.Format(time.RFC3339),
		"end_time":           task.EndTime.Format(time.RFC3339),
		"execution_duration": task.ExecutionDuration.String(),
	}
}

func UnmarshalTaskRedisStream(src map[string]interface{}) boltzmann.Task {
	startTime, _ := time.Parse(time.RFC3339, src["start_time"].(string))
	endTime, _ := time.Parse(time.RFC3339, src["end_time"].(string))
	execDuration, _ := time.ParseDuration(src["execution_duration"].(string))
	args := map[string]string{}
	argsJSON := []byte(src["agent_arguments"].(string))
	_ = jsoniter.Unmarshal(argsJSON, &args)
	return boltzmann.Task{
		TaskID:            src["task_id"].(string),
		CorrelationID:     src["correlation_id"].(string),
		Driver:            src["driver"].(string),
		ResourceURI:       src["resource_uri"].(string),
		AgentArguments:    args,
		Payload:           []byte(src["payload"].(string)),
		Status:            boltzmann.NewTaskStatus(src["status"].(string)),
		SuccessMessage:    src["success_message"].(string),
		FailureMessage:    src["failure_message"].(string),
		StartTime:         startTime,
		EndTime:           endTime,
		ExecutionDuration: execDuration,
	}
}
