package executor

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann/internal/executor/delegate"
)

// ConcurrentExecutor is the concurrency-backed implementation of Executor.
// This implementation executes tasks in different goroutines and reduces host's resource exhaustion by
// using a semaphore (can be configured with ConcurrentExecutorConfig).
type ConcurrentExecutor[T any] struct {
	Config   ConcurrentExecutorConfig
	Delegate delegate.Delegate[T]

	procSemaphoreLock sync.Mutex
	procSemaphore     *semaphore.Weighted
}

var _ Executor[string] = &ConcurrentExecutor[string]{}

func (s *ConcurrentExecutor[T]) ExecuteAll(ctx context.Context, args []T) error {
	s.procSemaphoreLock.Lock()
	if s.procSemaphore == nil {
		s.procSemaphore = semaphore.NewWeighted(s.Config.MaxGoroutines)
		s.procSemaphoreLock.Unlock()
	} else {
		s.procSemaphoreLock.Unlock()
	}

	errs := atomic.Pointer[multierror.Error]{}
	errs.Store(&multierror.Error{})
	for _, arg := range args {
		if err := s.procSemaphore.Acquire(ctx, 1); err != nil {
			return err
		}
		go func(processArg T) {
			defer s.procSemaphore.Release(1)
			if errExec := s.Delegate.Execute(ctx, processArg); errExec != nil {
				errs.Store(multierror.Append(errs.Load(), errExec))
			}
		}(arg)
	}

	return errs.Load().ErrorOrNil()
}
