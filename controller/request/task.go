package request

import "github.com/neutrinocorp/boltzmann/command"

type ScheduleTasksRequest struct {
	Tasks []command.ScheduleTaskCommand `json:"tasks"`
}
