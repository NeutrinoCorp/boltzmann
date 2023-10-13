package response

import (
	"time"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/scheduler"
)

type ScheduledTaskResponse struct {
	TaskID        string    `json:"task_id"`
	CorrelationID string    `json:"correlation_id"`
	Driver        string    `json:"driver"`
	ResourceURI   string    `json:"resource_uri"`
	ScheduleTime  time.Time `json:"schedule_time"`
}

type ScheduledTasksResponse struct {
	Tasks []ScheduledTaskResponse `json:"tasks"`
}

func NewScheduledTasksResponse(schedTasks []scheduler.ScheduleTaskResult) ScheduledTasksResponse {
	res := ScheduledTasksResponse{
		Tasks: make([]ScheduledTaskResponse, 0, len(schedTasks)),
	}
	for _, result := range schedTasks {
		res.Tasks = append(res.Tasks, ScheduledTaskResponse{
			TaskID:        result.TaskID,
			CorrelationID: result.CorrelationID,
			Driver:        result.Driver,
			ResourceURI:   result.ResourceURI,
			ScheduleTime:  result.ScheduleTime,
		})
	}

	return res
}

type TaskResponse struct {
	TaskID                  string            `json:"task_id"`
	CorrelationID           string            `json:"correlation_id"`
	Driver                  string            `json:"driver"`
	ResourceURI             string            `json:"resource_uri"`
	AgentArguments          map[string]string `json:"agent_arguments"`
	Payload                 []byte            `json:"payload"`
	Status                  string            `json:"status"`
	Response                string            `json:"response"`
	FailureMessage          string            `json:"failure_message"`
	ScheduleTime            time.Time         `json:"schedule_time"`
	StartTime               time.Time         `json:"start_time"`
	EndTime                 time.Time         `json:"end_time"`
	ExecutionDurationMillis int64             `json:"execution_duration_millis"`
}

type TaskContainerResponse struct {
	Task TaskResponse `json:"task"`
}

func NewContainerResponse(task boltzmann.Task) TaskContainerResponse {
	return TaskContainerResponse{
		Task: TaskResponse{
			TaskID:                  task.TaskID,
			CorrelationID:           task.CorrelationID,
			Driver:                  task.Driver,
			ResourceURI:             task.ResourceURI,
			AgentArguments:          task.AgentArguments,
			Payload:                 task.Payload,
			Status:                  task.Status.String(),
			ScheduleTime:            task.ScheduleTime,
			Response:                string(task.Response),
			FailureMessage:          task.FailureMessage,
			StartTime:               task.StartTime,
			EndTime:                 task.EndTime,
			ExecutionDurationMillis: task.ExecutionDuration.Milliseconds(),
		},
	}
}
