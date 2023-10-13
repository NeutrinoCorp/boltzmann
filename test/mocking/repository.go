package mocking

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/neutrinocorp/boltzmann"
)

type StateRepository struct {
	mock.Mock
}

func (s *StateRepository) Save(ctx context.Context, task boltzmann.Task) error {
	args := s.Called(ctx, task)
	return args.Error(0)
}

func (s *StateRepository) SaveAll(ctx context.Context, tasks ...boltzmann.Task) error {
	args := s.Called(ctx, tasks)
	return args.Error(0)
}

func (s *StateRepository) Get(ctx context.Context, taskId string) (boltzmann.Task, error) {
	args := s.Called(ctx, taskId)
	return args.Get(0).(boltzmann.Task), args.Error(1)
}
