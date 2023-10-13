package state_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/state"
	"github.com/neutrinocorp/boltzmann/test/mocking"
)

type serviceSuite struct {
	suite.Suite

	idFactory *mocking.FakeIDFactory
	repo      *mocking.StateRepository
	svc       state.Service
}

func TestService(t *testing.T) {
	suite.Run(t, &serviceSuite{})
}

func (s *serviceSuite) SetupSuite() {
	s.idFactory = &mocking.FakeIDFactory{}
	s.repo = &mocking.StateRepository{}
	s.svc = state.Service{
		StateRepository: s.repo,
	}
}

func (s *serviceSuite) Test_Get() {
	ctx := context.Background()
	taskID := "123"
	s.repo.On("Get", ctx, taskID).Return(boltzmann.Task{
		TaskID: taskID,
	}, error(nil))
	out, err := s.svc.Get(ctx, taskID)
	s.Assert().NoError(err)
	s.Assert().Equal(taskID, out.TaskID)
	s.repo.AssertCalled(s.T(), "Get", ctx, taskID)
}
