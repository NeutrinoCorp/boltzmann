package queue

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"golang.org/x/sync/semaphore"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/marshal"
)

type RedisServiceConfig struct {
	StreamName                    string
	StreamGroupID                 string
	MaxInFlightProcesses          int64
	EnableStreamGroupAutoCreation bool
	RetryBackoff                  time.Duration
}

type RedisService struct {
	Client *redis.Client
	Config RedisServiceConfig

	consumerID               string
	baseCtx                  context.Context
	baseCtxCancelFunc        context.CancelFunc
	streamReaderSemaphore    *semaphore.Weighted
	inFlightGroupLock        *sync.WaitGroup // required to avoid race conditions
	inFlightStreamReaderLock *sync.WaitGroup
}

var _ Service = RedisService{}

func NewRedisService(c *redis.Client, cfg RedisServiceConfig) RedisService {
	return RedisService{
		Client:                   c,
		Config:                   cfg,
		consumerID:               xid.New().String(),
		streamReaderSemaphore:    semaphore.NewWeighted(cfg.MaxInFlightProcesses),
		inFlightGroupLock:        &sync.WaitGroup{},
		inFlightStreamReaderLock: &sync.WaitGroup{},
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

	r.baseCtx, r.baseCtxCancelFunc = context.WithCancel(ctx)
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
		Block:    0,
		NoAck:    false,
	}).Result()
	if err != nil || len(stream) == 0 {
		redisSvcLogger.Err(err).
			Str("stream_name", r.Config.StreamName).
			Str("group_id", r.Config.StreamGroupID).
			Str("consumer_id", r.consumerID).
			Msg("cannot read from stream")
		time.Sleep(r.Config.RetryBackoff)
		return
	}

	if err = r.processStream(ctx, stream[0].Messages); err != nil {
		redisSvcLogger.Err(err).
			Str("stream_name", r.Config.StreamName).
			Str("group_id", r.Config.StreamGroupID).
			Str("consumer_id", r.consumerID).
			Msg("failed to process stream")
		time.Sleep(r.Config.RetryBackoff)
		return
	}

	redisSvcLogger.Info().
		Str("stream_name", r.Config.StreamName).
		Str("group_id", r.Config.StreamGroupID).
		Str("consumer_id", r.consumerID).
		Msg("successfully processed stream")
}

func (r RedisService) processStream(ctx context.Context, messages []redis.XMessage) error {
	r.inFlightStreamReaderLock.Add(len(messages))

	failedMsgIDBufferAtomic := atomic.Value{}
	failedMsgIDBufferAtomic.Store(make([]string, 0, len(messages)))
	for _, msg := range messages {
		if err := r.streamReaderSemaphore.Acquire(ctx, 1); err != nil {
			r.inFlightStreamReaderLock.Done()
			return err
		}
		go r.execAgent(ctx, failedMsgIDBufferAtomic, msg.ID, marshal.UnmarshalTaskRedisStream(msg.Values))
	}
	r.inFlightStreamReaderLock.Wait()

	failedMsgIDBuffer := failedMsgIDBufferAtomic.Load().([]string)
	return r.Client.XAck(ctx, r.Config.StreamName, r.Config.StreamGroupID, failedMsgIDBuffer...).Err()
}

func (r RedisService) execAgent(_ context.Context, failedMsgIDBuffer atomic.Value, msgID string, task boltzmann.Task) {
	defer r.inFlightStreamReaderLock.Done()
	defer r.streamReaderSemaphore.Release(1)
	var err error
	defer func() {
		if err == nil {
			return
		}
		failedMsgIDBuffer.Store(append(failedMsgIDBuffer.Load().([]string), msgID))
	}()

	// TODO: Perform actual agent
	redisSvcLogger.Info().
		Str("task", fmt.Sprintf("%+v", task)).
		Msg("executing agent...")
}

func (r RedisService) Shutdown(_ context.Context) error {
	r.inFlightStreamReaderLock.Wait()
	r.inFlightGroupLock.Wait()
	r.baseCtxCancelFunc()
	return nil
}
