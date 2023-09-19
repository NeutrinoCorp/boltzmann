package queue

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/marshal"
	"github.com/neutrinocorp/boltzmann/state"
)

type RedisServiceConfig struct {
	StreamName                    string
	StreamGroupID                 string
	MaxInFlightProcesses          int64
	EnableStreamGroupAutoCreation bool
	RetryBackoff                  time.Duration
}

type RedisService struct {
	Client          *redis.Client
	Config          RedisServiceConfig
	AgentRegistry   agent.Registry
	StateRepository state.Repository

	consumerID               string
	baseCtx                  context.Context
	baseCtxCancelFunc        context.CancelFunc
	streamReaderSemaphore    *semaphore.Weighted
	inFlightGroupLock        *sync.WaitGroup // required to avoid race conditions
	inFlightStreamReaderLock *sync.WaitGroup
}

var _ Service = RedisService{}

func NewRedisService(c *redis.Client, cfg RedisServiceConfig, agentReg agent.Registry, stateRepo state.RedisRepository) RedisService {
	ctx, cancel := context.WithCancel(context.Background())
	return RedisService{
		Client:                   c,
		Config:                   cfg,
		AgentRegistry:            agentReg,
		StateRepository:          stateRepo,
		consumerID:               xid.New().String(),
		streamReaderSemaphore:    semaphore.NewWeighted(cfg.MaxInFlightProcesses),
		inFlightGroupLock:        &sync.WaitGroup{},
		inFlightStreamReaderLock: &sync.WaitGroup{},
		baseCtx:                  ctx,
		baseCtxCancelFunc:        cancel,
	}
}

func (r RedisService) Enqueue(ctx context.Context, task boltzmann.Task) error {
	return r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream:     r.Config.StreamName,
		NoMkStream: false,
		MaxLen:     0,
		MinID:      "",
		Approx:     false,
		Limit:      0,
		ID:         "*",
		Values:     marshal.MarshalTaskRedisStream(task),
	}).Err()
}

func (r RedisService) Start(ctx context.Context) error {
	if err := r.ensureStreamGroup(ctx); err != nil {
		return err
	}

consumerLoop:
	for {
		select {
		case <-r.baseCtx.Done():
			break consumerLoop
		default:
		}

		redisSvcLogger.Info().Msg("polling messages")
		r.inFlightGroupLock.Add(1)
		go r.readStream(r.baseCtx)
		r.inFlightGroupLock.Wait()
	}
	redisSvcLogger.Info().Msg("closed stream reader")
	return nil
}

func (r RedisService) ensureStreamGroup(ctx context.Context) error {
	err := r.Client.XGroupCreate(ctx, r.Config.StreamName, r.Config.StreamGroupID, "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil
}

func (r RedisService) readStream(ctx context.Context) {
	defer r.inFlightGroupLock.Done()
	stream, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    r.Config.StreamGroupID,
		Consumer: r.consumerID,
		Streams:  []string{r.Config.StreamName, ">"},
		Count:    r.Config.MaxInFlightProcesses,
		Block:    r.Config.RetryBackoff,
		NoAck:    false,
	}).Result()
	if err != nil && !(err.Error() == "redis: nil") {
		redisSvcLogger.Err(err).
			Str("stream_name", r.Config.StreamName).
			Str("group_id", r.Config.StreamGroupID).
			Str("consumer_id", r.consumerID).
			Msg("cannot read from stream")
		return
	} else if len(stream) == 0 {
		redisSvcLogger.Warn().
			Str("stream_name", r.Config.StreamName).
			Str("group_id", r.Config.StreamGroupID).
			Str("consumer_id", r.consumerID).
			Msg("empty stream")
		return
	}

	if err = r.processStream(ctx, stream[0].Messages); err != nil {
		redisSvcLogger.Err(err).
			Str("stream_name", r.Config.StreamName).
			Str("group_id", r.Config.StreamGroupID).
			Str("consumer_id", r.consumerID).
			Msg("failed to process stream")
		return
	}

	redisSvcLogger.Info().
		Str("stream_name", r.Config.StreamName).
		Str("group_id", r.Config.StreamGroupID).
		Str("consumer_id", r.consumerID).
		Msg("successfully processed stream")
}

func (r RedisService) processStream(ctx context.Context, messages []redis.XMessage) error {
	succeedMsgIDBufferAtomic := atomic.Value{}
	succeedMsgIDBufferAtomic.Store(make([]string, 0, len(messages)))

	r.inFlightStreamReaderLock.Add(len(messages))
	for _, msg := range messages {
		if err := r.streamReaderSemaphore.Acquire(ctx, 1); err != nil {
			r.inFlightStreamReaderLock.Done()
			return err
		}
		go r.execAgent(ctx, &succeedMsgIDBufferAtomic, msg)
	}
	r.inFlightStreamReaderLock.Wait()

	succeedMsgIDBuffer := succeedMsgIDBufferAtomic.Load().([]string)
	return r.Client.XAck(ctx, r.Config.StreamName, r.Config.StreamGroupID, succeedMsgIDBuffer...).Err()
}

func (r RedisService) execAgent(_ context.Context, succeedMsgIDBuffer *atomic.Value,
	msg redis.XMessage) {
	defer r.streamReaderSemaphore.Release(1)
	defer r.inFlightStreamReaderLock.Done()

	task := marshal.UnmarshalTaskRedisStream(msg.Values)
	ctx, cancel := context.WithTimeout(r.baseCtx, time.Second*120)
	defer cancel()
	var err error
	defer func() {
		task.Status = boltzmann.TaskStatusSucceed
		if err != nil {
			task.Status = boltzmann.TaskStatusFailed
			task.FailureMessage = err.Error()
		} else {
			succeedMsgIDBuffer.Store(append(succeedMsgIDBuffer.Load().([]string), msg.ID))
		}

		task.EndTime = time.Now().UTC()
		task.ExecutionDuration = task.EndTime.Sub(task.StartTime)
		if errCommit := r.StateRepository.Save(ctx, task); errCommit != nil {
			embeddedSvcLogger.Err(errCommit).
				Str("task_id", task.TaskID).
				Str("driver", task.Driver).
				Str("resource_location", task.ResourceURI).
				Msg("failed to save state")
		}
	}()

	redisSvcLogger.Info().
		Str("task_id", task.TaskID).
		Msg("executing agent...")

	task.Status = boltzmann.TaskStatusPending
	if errCommit := r.StateRepository.Save(ctx, task); errCommit != nil {
		embeddedSvcLogger.Err(errCommit).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to save state")
		return
	}

	taskAgent, errAgent := r.AgentRegistry.Get(task.Driver)
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

func (r RedisService) Shutdown(_ context.Context) error {
	r.inFlightStreamReaderLock.Wait()
	r.inFlightGroupLock.Wait()
	r.baseCtxCancelFunc()
	return nil
}
