package queue

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/config"
)

var serviceLogger = log.With().Str("component", "queue.service").Logger()

type ServiceConfig struct {
	FetchInterval time.Duration
	RetryInterval time.Duration
	JobTimeout    time.Duration
	MaxRetries    int8
	MaxProc       int64
}

func setServiceConfigDefault() {
	config.SetDefault(config.QueueFetchInterval, time.Second*3)
	config.SetDefault(config.QueueRetryInterval, time.Second*5)
	config.SetDefault(config.QueueJobTimeout, time.Second*60)
	config.SetDefault(config.QueueMaxRetry, int8(-1))
	config.SetDefault(config.QueueMaxProc, int64(10))
}

func NewServiceConfig() ServiceConfig {
	setServiceConfigDefault()
	return ServiceConfig{
		FetchInterval: config.Get[time.Duration](config.QueueFetchInterval),
		RetryInterval: config.Get[time.Duration](config.QueueRetryInterval),
		JobTimeout:    config.Get[time.Duration](config.QueueJobTimeout),
		MaxRetries:    config.Get[int8](config.QueueMaxRetry),
		MaxProc:       config.Get[int64](config.QueueMaxProc),
	}
}

type Service struct {
	Config        ServiceConfig
	Queue         Queue
	AgentRegistry agent.Registry

	baseCtx           context.Context
	baseCtxCancelFunc context.CancelFunc
	inFlightWg        sync.WaitGroup
	semaphore         *semaphore.Weighted
}

func NewService(cfg ServiceConfig, agentReg agent.Registry, queue Queue) *Service {
	return &Service{
		Config:            cfg,
		Queue:             queue,
		AgentRegistry:     agentReg,
		baseCtx:           nil,
		baseCtxCancelFunc: nil,
		inFlightWg:        sync.WaitGroup{},
		semaphore:         semaphore.NewWeighted(cfg.MaxProc),
	}
}

func (d *Service) Start(ctx context.Context) error {
	d.baseCtx, d.baseCtxCancelFunc = context.WithCancel(ctx)
	retryCount := int8(0)
mainLoop:
	for {
		select {
		case <-ctx.Done():
			break mainLoop
		default:
		}

		d.inFlightWg.Add(1)
		serviceLogger.Info().Msg("fetching tasks from queue")
		tasks, err := d.Queue.Pop(d.baseCtx)
		if err != nil {
			serviceLogger.Err(err).Msg("failed to fetch tasks from queue")
			retryCount++
			if d.Config.MaxRetries >= 0 && retryCount >= d.Config.MaxRetries {
				break // stop process after max retries reached
			}
			time.Sleep(d.Config.RetryInterval)
			continue
		}
		d.execTasks(tasks)
		d.inFlightWg.Done()
		time.Sleep(d.Config.FetchInterval)
	}

	return nil
}

func (d *Service) execTasks(tasks []boltzmann.Task) {
	d.inFlightWg.Add(len(tasks))
	if err := d.semaphore.Acquire(d.baseCtx, int64(len(tasks))); err != nil {
		serviceLogger.Err(err).Msg("cannot acquire semaphore")
		return
	}
	for i := 0; i < len(tasks); i++ {
		go d.execTask(tasks[i])
	}
}

func (d *Service) execTask(task boltzmann.Task) {
	defer d.inFlightWg.Done()
	defer d.semaphore.Release(1)
	// observability is available through agent middlewares
	taskAgent, err := d.AgentRegistry.Get(task.Driver)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(d.baseCtx, d.Config.JobTimeout)
	defer cancel()
	_, _ = taskAgent.Execute(ctx, task)
}

func (d *Service) Shutdown(_ context.Context) error {
	serviceLogger.Info().Msg("shutting down service")
	d.inFlightWg.Wait()
	d.baseCtxCancelFunc()
	serviceLogger.Info().Msg("gracefully shut down service")
	return nil
}
