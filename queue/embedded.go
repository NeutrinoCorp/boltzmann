package queue

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/state"
)

type EmbeddedServiceConfig struct {
	BufferSize           int
	MaxInFlightProcesses int64
}

// EmbeddedService is the Service queuing service which uses goroutines and channels to operate.
type EmbeddedService struct {
	AgentRegistry   agent.Registry
	StateRepository state.Repository

	sem           *semaphore.Weighted
	messageBuffer chan boltzmann.Task
	inFlightLock  sync.WaitGroup
}

func NewEmbeddedService(cfg EmbeddedServiceConfig, registry agent.Registry, state state.Repository) *EmbeddedService {
	return &EmbeddedService{
		AgentRegistry:   registry,
		StateRepository: state,
		sem:             semaphore.NewWeighted(cfg.MaxInFlightProcesses),
		messageBuffer:   make(chan boltzmann.Task, cfg.BufferSize<<0),
		inFlightLock:    sync.WaitGroup{},
	}
}

var _ Service = &EmbeddedService{}

func (s *EmbeddedService) Start(ctx context.Context) error {
	for task := range s.messageBuffer {
		if err := s.sem.Acquire(ctx, 1); err != nil {
			continue
		}
		// this is running inside a new goroutine (as we are listening to a channel)
		go s.execAgent(ctx, task)
	}
	return nil
}

func (s *EmbeddedService) execAgent(rootCtx context.Context, task boltzmann.Task) {
	defer s.inFlightLock.Done()
	defer s.sem.Release(1)

	task.Status = boltzmann.TaskStatusPending
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*120)
	defer cancel()
	if errCommit := s.StateRepository.Save(ctx, task); errCommit != nil {
		embeddedSvcLogger.Err(errCommit).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to save state")
		return
	}

	var err error
	defer func() {
		task.Status = boltzmann.TaskStatusSucceed
		if err != nil {
			task.Status = boltzmann.TaskStatusFailed
			task.FailureMessage = err.Error()
		}

		task.EndTime = time.Now().UTC()
		task.ExecutionDuration = task.EndTime.Sub(task.StartTime)
		if errCommit := s.StateRepository.Save(ctx, task); errCommit != nil {
			embeddedSvcLogger.Err(errCommit).
				Str("task_id", task.TaskID).
				Str("driver", task.Driver).
				Str("resource_location", task.ResourceURI).
				Msg("failed to save state")
		}
	}()

	taskAgent, errAgent := s.AgentRegistry.Get(task.Driver)
	if errAgent != nil {
		embeddedSvcLogger.Err(errAgent).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to execute task")
		err = errAgent
		return
	}

	err = taskAgent.Execute(ctx, task)
	if err != nil {
		embeddedSvcLogger.Err(err).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to execute task")
	}
}

func (s *EmbeddedService) Shutdown(_ context.Context) error {
	log.Info().Msg("gracefully shutting down")
	s.inFlightLock.Wait()
	close(s.messageBuffer)
	log.Info().Msg("service has been shut down")
	return nil
}

func (s *EmbeddedService) Enqueue(_ context.Context, task boltzmann.Task) error {
	s.inFlightLock.Add(1)
	s.messageBuffer <- task
	return nil
}
