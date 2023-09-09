package marshal

import (
	"time"

	"github.com/neutrinocorp/boltzmann"
)

func MarshalTaskRedisStream(task boltzmann.Task) map[string]interface{} {
	return map[string]interface{}{
		"task_id":            task.TaskID,
		"correlation_id":     task.CorrelationID,
		"driver":             task.Driver,
		"resource_uri":       task.ResourceURI,
		"agent_arguments":    task.AgentArguments,
		"payload":            task.Payload,
		"status":             task.Status.String(),
		"success_message":    task.SuccessMessage,
		"failure_message":    task.FailureMessage,
		"start_time":         task.StartTime.String(),
		"end_time":           task.EndTime.String(),
		"execution_duration": task.ExecutionDuration.String(),
	}
}

func UnmarshalTaskRedisStream(src map[string]interface{}) boltzmann.Task {
	startTime, _ := time.Parse(time.RFC3339, src["start_time"].(string))
	endTime, _ := time.Parse(time.RFC3339, src["end_time"].(string))
	execDuration, _ := time.ParseDuration(src["execution_duration"].(string))
	return boltzmann.Task{
		TaskID:            src["task_id"].(string),
		CorrelationID:     src["correlation_id"].(string),
		Driver:            src["driver"].(string),
		ResourceURI:       src["resource_uri"].(string),
		AgentArguments:    src["agent_arguments"].(map[string]string),
		Payload:           src["payload"].([]byte),
		Status:            boltzmann.NewTaskStatus(src["status"].(string)),
		SuccessMessage:    src["success_message"].(string),
		FailureMessage:    src["failure_message"].(string),
		StartTime:         startTime,
		EndTime:           endTime,
		ExecutionDuration: execDuration,
	}
}
