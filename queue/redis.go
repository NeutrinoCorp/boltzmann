package queue

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/config"
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

func NewRedisServiceConfig() RedisServiceConfig {
	config.SetDefault("STREAM_NAME", "boltzmann-job-queue")
	config.SetDefault("STREAM_GROUP_ID", "boltzmann-agent-worker_pool")
	config.SetDefault("ENABLE_STREAM_GROUP_AUTO_CREATE", true)
	config.SetDefault("RETRY_BACKOFF", time.Second*3)
	config.SetDefault("MAX_IN_FLIGHT_PROCESSES", int64(runtime.GOMAXPROCS(0)))
	return RedisServiceConfig{
		StreamName:                    config.GetEnv[string]("STREAM_NAME"),
		StreamGroupID:                 config.GetEnv[string]("STREAM_GROUP_ID"),
		MaxInFlightProcesses:          config.GetEnv[int64]("MAX_IN_FLIGHT_PROCESSES"),
		EnableStreamGroupAutoCreation: config.GetEnv[bool]("ENABLE_STREAM_GROUP_AUTO_CREATE"),
		RetryBackoff:                  config.GetEnv[time.Duration]("RETRY_BACKOFF"),
	}
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
		ID:         "",
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
			redisSvcLogger.Info().Msg("closing stream reader")
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
	msgIdBuffer := make([]string, 0, len(messages))
	// supervisor will enqueue failed tasks, thus, ack ALL messages to avoid infinite loops
	defer func() {
		redisSvcLogger.Info().
			Int("total_messages", len(msgIdBuffer)).
			Msg("committed messages")
	}()
	r.inFlightStreamReaderLock.Add(len(messages))
	for _, msg := range messages {
		msgIdBuffer = append(msgIdBuffer, msg.ID)
		if err := r.streamReaderSemaphore.Acquire(ctx, 1); err != nil {
			r.inFlightStreamReaderLock.Done()
			return err
		}
		go r.execAgent(ctx, msg)
	}
	r.inFlightStreamReaderLock.Wait()

	return r.Client.XAck(ctx, r.Config.StreamName, r.Config.StreamGroupID, msgIdBuffer...).Err()
}

func (r RedisService) execAgent(_ context.Context, msg redis.XMessage) {
	defer r.streamReaderSemaphore.Release(1)
	defer r.inFlightStreamReaderLock.Done()

	task := marshal.UnmarshalTaskRedisStream(msg.Values)
	ctx, cancel := context.WithTimeout(r.baseCtx, time.Second*120)
	defer cancel()
	var err error
	defer func() {
		if err == nil {
			return
		}

		redisSvcLogger.Err(err).
			Str("task_id", task.TaskID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to execute task")
	}()

	redisSvcLogger.Info().
		Str("task_id", task.TaskID).
		Msg("executing agent...")

	taskAgent, errAgent := r.AgentRegistry.Get(task.Driver)
	if errAgent != nil {
		err = errAgent
		return
	}

	err = taskAgent.Execute(ctx, task)
}

func (r RedisService) Shutdown(_ context.Context) error {
	r.inFlightStreamReaderLock.Wait()
	r.inFlightGroupLock.Wait()
	redisSvcLogger.Info().Msg("shutting down stream reader")
	r.baseCtxCancelFunc()
	return nil
}
