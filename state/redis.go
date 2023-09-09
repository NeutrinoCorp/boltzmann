package state

import (
	"context"
	"fmt"
	"time"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/codec"

	"github.com/redis/go-redis/v9"
)

const redisKeyPattern = "task::%s"

type RedisRepository struct {
	Client *redis.Client
	Codec  codec.Codec
}

var _ Repository = RedisRepository{}

func (r RedisRepository) Save(ctx context.Context, task boltzmann.Task) error {
	key := fmt.Sprintf(redisKeyPattern, task.TaskID)

	encodedTask, err := r.Codec.Encode(task)
	if err != nil {
		return err
	}

	cmd := r.Client.Set(ctx, key, encodedTask, time.Hour*72)
	return cmd.Err()
}

func (r RedisRepository) Get(ctx context.Context, taskId string) (boltzmann.Task, error) {
	cmd := r.Client.Get(ctx, fmt.Sprintf(redisKeyPattern, taskId))
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
