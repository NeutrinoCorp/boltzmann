package boltzmann

import (
	"fmt"
)

const (
	_ TaskStatus = iota
	TaskStatusScheduled
	TaskStatusStarted
	TaskStatusFailed
	TaskStatusSucceed
)

var taskStatusMap = map[TaskStatus]string{
	TaskStatusScheduled: "SCHEDULED",
	TaskStatusStarted:   "STARTED",
	TaskStatusFailed:    "FAILED",
	TaskStatusSucceed:   "SUCCEED",
}

var taskStatusMapBackward = map[string]TaskStatus{
	"SCHEDULED": TaskStatusScheduled,
	"STARTED":   TaskStatusStarted,
	"FAILED":    TaskStatusFailed,
	"SUCCEED":   TaskStatusSucceed,
}

type TaskStatus uint8

var _ fmt.Stringer = TaskStatus(0)

var _ fmt.GoStringer = TaskStatus(0)

func NewTaskStatus(status string) TaskStatus {
	return taskStatusMapBackward[status]
}

func (s TaskStatus) String() string {
	return taskStatusMap[s]
}

func (s TaskStatus) GoString() string {
	return s.String()
}
