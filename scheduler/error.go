package scheduler

import "github.com/neutrinocorp/boltzmann"

type SchedulingError struct {
	Task  boltzmann.Task
	Cause error
}

var _ error = SchedulingError{}

func (s SchedulingError) Error() string {
	return s.Cause.Error()
}
