package state

import (
	"context"
	"sync"

	"github.com/neutrinocorp/boltzmann"
)

type Repository interface {
	Save(ctx context.Context, task boltzmann.Task) error
	Get(ctx context.Context, taskId string) (boltzmann.Task, error)
}

type EmbeddedRepository struct {
	db map[string]boltzmann.Task
	mu sync.RWMutex
}

func NewEmbeddedRepository() *EmbeddedRepository {
	return &EmbeddedRepository{
		db: make(map[string]boltzmann.Task),
		mu: sync.RWMutex{},
	}
}

var _ Repository = &EmbeddedRepository{}

func (s *EmbeddedRepository) Save(_ context.Context, task boltzmann.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db[task.TaskID] = task
	return nil
}

func (s *EmbeddedRepository) Get(_ context.Context, taskId string) (boltzmann.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.db[taskId]
	if !ok {
		return boltzmann.Task{}, ErrTaskStateNotFound
	}

	return task, nil
}
