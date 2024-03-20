package state

import (
	"context"
	"fmt"
	"time"

	"github.com/modern-go/reflect2"
)

type Service[T any] struct {
	Repository Repository
}

func (s Service[T]) Create(ctx context.Context, resourceID string) error {
	var zeroVal T
	typeOf := reflect2.TypeOf(zeroVal).String()
	key := fmt.Sprintf("%s#%s", typeOf, resourceID)
	state := State{
		ID:           key,
		ResourceName: typeOf,
		ResourceID:   resourceID,
		Status:       StatusRunning,
		StartTime:    time.Now().UTC(),
	}
	return s.Repository.Save(ctx, state)
}

func (s Service[T]) MarkAsFailed(ctx context.Context, resourceID string, procErr error) error {
	state, err := s.Get(ctx, resourceID)
	if err != nil {
		return err
	}

	state.Status = StatusFailed
	state.ExecutionError = procErr.Error()
	state.EndTime = time.Now().UTC()
	state.ExecutionDuration = state.EndTime.Sub(state.StartTime)
	state.ExecutionDurationMillis = state.ExecutionDuration.Milliseconds()
	return s.Repository.Save(ctx, state)
}

func (s Service[T]) MarkAsCompleted(ctx context.Context, resourceID string) error {
	state, err := s.Get(ctx, resourceID)
	if err != nil {
		return err
	}

	state.Status = StatusCompleted
	state.ExecutionError = ""
	state.EndTime = time.Now().UTC()
	state.ExecutionDuration = state.EndTime.Sub(state.StartTime)
	state.ExecutionDurationMillis = state.ExecutionDuration.Milliseconds()
	return s.Repository.Save(ctx, state)
}

func (s Service[T]) Get(ctx context.Context, resourceID string) (State, error) {
	var zeroVal T
	typeOf := reflect2.TypeOf(zeroVal).String()
	key := fmt.Sprintf("%s#%s", typeOf, resourceID)
	return s.Repository.GetByID(ctx, key)
}
