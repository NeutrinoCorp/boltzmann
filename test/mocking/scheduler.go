package mocking

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/scheduler"
)

type Scheduler struct {
	mock.Mock
}

var _ scheduler.TaskScheduler = &Scheduler{}

func (s *Scheduler) Schedule(ctx context.Context, _ []boltzmann.Task) ([]scheduler.ScheduleTaskResult, error) {
	args := s.Called(ctx, mock.AnythingOfType("[]boltzmann.Task"))
	return args.Get(0).([]scheduler.ScheduleTaskResult), args.Error(1)
}
