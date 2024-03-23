package state

import (
	"context"
)

type Service[T any] struct {
	Repository Repository
}

func (s Service[T]) Create(ctx context.Context, resourceID string) error {
	state := createState[T](resourceID)
	return s.Repository.Save(ctx, state)
}

func (s Service[T]) MarkAsFailed(ctx context.Context, resourceID string, procErr error) error {
	state, err := s.Get(ctx, resourceID)
	if err != nil {
		return err
	}

	state.markAsFailed(procErr)
	return s.Repository.Save(ctx, state)
}

func (s Service[T]) MarkAsCompleted(ctx context.Context, resourceID string) error {
	state, err := s.Get(ctx, resourceID)
	if err != nil {
		return err
	}

	state.markAsCompleted()
	return s.Repository.Save(ctx, state)
}

func (s Service[T]) Get(ctx context.Context, resourceID string) (State, error) {
	key := newStateID[T](resourceID)
	return s.Repository.GetByID(ctx, key)
}
