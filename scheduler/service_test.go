package scheduler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/command"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/test/mocking"
)

type serviceSuite struct {
	suite.Suite

	idFactory *mocking.FakeIDFactory
	repo      *mocking.StateRepository
	sched     *mocking.Scheduler
	svc       scheduler.Service
}

func TestService(t *testing.T) {
	suite.Run(t, &serviceSuite{})
}

func (s *serviceSuite) SetupSuite() {
	s.idFactory = &mocking.FakeIDFactory{}
	s.repo = &mocking.StateRepository{}
	s.sched = &mocking.Scheduler{}
	s.svc = scheduler.Service{
		Scheduler:       s.sched,
		StateRepository: s.repo,
		FactoryID:       s.idFactory,
	}
}

func (s *serviceSuite) Test_GetTaskState() {
	ctx := context.Background()
	taskID := "123"
	s.repo.On("Get", ctx, taskID).Return(boltzmann.Task{
		TaskID: taskID,
	}, error(nil))
	out, err := s.svc.GetTaskState(ctx, taskID)
	s.Assert().NoError(err)
	s.Assert().Equal(taskID, out.TaskID)
	s.repo.AssertCalled(s.T(), "Get", ctx, taskID)
}

func (s *serviceSuite) Test_Schedule() {
	ctx := context.Background()
	commands := []command.ScheduleTaskCommand{
		{
			Driver:      "foo",
			ResourceURI: "foo_1",
		},
		{
			Driver:      "bar",
			ResourceURI: "bar_1",
		},
		{
			Driver:      "baz",
			ResourceURI: "baz_1",
		},
	}
	results := []scheduler.ScheduleTaskResult{
		{
			TaskID:      "123",
			Driver:      "foo",
			ResourceURI: "foo_1",
		},
	}

	s.idFactory.On("NewID").Return("123", error(nil))
	s.sched.On("Schedule", ctx, mock.Anything).Return(results, error(nil))
	out, err := s.svc.Schedule(ctx, commands)
	s.Assert().NoError(err)

	s.sched.AssertCalled(s.T(), "Schedule", ctx, mock.Anything)
	s.Assert().Equal(results[0].TaskID, out[0].TaskID)
	s.Assert().Equal(results[0].Driver, out[0].Driver)
	s.Assert().Equal(results[0].ResourceURI, out[0].ResourceURI)
}
