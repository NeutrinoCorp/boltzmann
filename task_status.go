package boltzmann

import "fmt"

const (
	_ TaskStatus = iota
	TaskStatusInit
	TaskStatusScheduled
	TaskStatusPending
	TaskStatusFailed
	TaskStatusSucceed
)

var taskStatusMap = map[TaskStatus]string{
	TaskStatusInit:      "INITIATED",
	TaskStatusScheduled: "SCHEDULED",
	TaskStatusPending:   "PENDING",
	TaskStatusFailed:    "FAILED",
	TaskStatusSucceed:   "SUCCEED",
}

var taskStatusMapBackward = map[string]TaskStatus{
	"INITIATED": TaskStatusInit,
	"SCHEDULED": TaskStatusScheduled,
	"PENDING":   TaskStatusPending,
	"FAILED":    TaskStatusFailed,
	"SUCCEED":   TaskStatusFailed,
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
