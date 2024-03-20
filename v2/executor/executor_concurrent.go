package executor

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann/v2/executor/delegate"
)

type ConcurrentExecutorConfig struct {
	MaxGoroutines int64
}

type ConcurrentExecutor[T any] struct {
	Config   ConcurrentExecutorConfig
	Delegate delegate.Delegate[T]

	procSemaphoreLock sync.Mutex
	procSemaphore     *semaphore.Weighted
}

var _ Executor[string] = &ConcurrentExecutor[string]{}

func (s *ConcurrentExecutor[T]) ExecuteAll(ctx context.Context, tasks []T) error {
	s.procSemaphoreLock.Lock()
	if s.procSemaphore == nil {
		s.procSemaphore = semaphore.NewWeighted(s.Config.MaxGoroutines)
		s.procSemaphoreLock.Unlock()
	} else {
		s.procSemaphoreLock.Unlock()
	}

	errs := atomic.Pointer[multierror.Error]{}
	errs.Store(&multierror.Error{})
	for _, task := range tasks {
		if err := s.procSemaphore.Acquire(ctx, 1); err != nil {
			return err
		}
		go func(procTask T) {
			defer s.procSemaphore.Release(1)
			if errExec := s.Delegate.Execute(ctx, procTask); errExec != nil {
				errs.Store(multierror.Append(errs.Load(), errExec))
			}
		}(task)
	}

	return errs.Load().ErrorOrNil()
}
