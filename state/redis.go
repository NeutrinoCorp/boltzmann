package state

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/codec"
	"github.com/neutrinocorp/boltzmann/config"
)

const redisKeyPrefix = "task"

func newRedisKey(prefix, key string) string {
	return prefix + "::" + key
}

type RedisRepositoryConfig struct {
	ItemTTL time.Duration
}

func setRedisConfigDefault() {
	config.SetDefault(config.RedisItemTTL, time.Hour*72)
}

func NewRedisRepositoryConfig() RedisRepositoryConfig {
	setRedisConfigDefault()
	return RedisRepositoryConfig{
		ItemTTL: config.Get[time.Duration](config.RedisItemTTL),
	}
}

type RedisRepository struct {
	Client *redis.Client
	Config RedisRepositoryConfig
	Codec  codec.Codec
}

var _ Repository = RedisRepository{}

func (r RedisRepository) Save(ctx context.Context, task boltzmann.Task) error {
	key := newRedisKey(redisKeyPrefix, task.TaskID)
	encodedTask, err := r.Codec.Encode(task)
	if err != nil {
		return err
	}

	cmd := r.Client.Set(ctx, key, encodedTask, r.Config.ItemTTL)
	return cmd.Err()
}

func (r RedisRepository) SaveAll(ctx context.Context, tasks ...boltzmann.Task) error {
	_, err := r.Client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		errs := &multierror.Error{}
		for _, task := range tasks {
			key := newRedisKey(redisKeyPrefix, task.TaskID)
			encodedTask, err := r.Codec.Encode(task)
			if err != nil {
				return err
			}

			if err = pipe.Set(ctx, key, encodedTask, r.Config.ItemTTL).Err(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
		return errs.ErrorOrNil()
	})
	return err
}

func (r RedisRepository) Get(ctx context.Context, taskId string) (boltzmann.Task, error) {
	key := newRedisKey(redisKeyPrefix, taskId)
	cmd := r.Client.Get(ctx, key)
	res, err := cmd.Result()
	if err != nil && err.Error() == "redis: nil" {
		return boltzmann.Task{}, ErrTaskStateNotFound
	} else if err != nil {
		return boltzmann.Task{}, err
	}

	task := boltzmann.Task{}
	if err = r.Codec.Decode([]byte(res), &task); err != nil {
		return boltzmann.Task{}, err
	}

	return task, err
}
