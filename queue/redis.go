package queue

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/codec"
	"github.com/neutrinocorp/boltzmann/config"
)

type RedisListConfig struct {
	QueueName string
	BatchSize int64
	IsLIFO    bool
}

func setRedisListConfigDefault() {
	config.SetDefault(config.QueueName, "boltzmann-job-queue")
	config.SetDefault(config.QueueBatchSize, int64(20))
	config.SetDefault(config.RedisEnableLIFO, false)
}

func NewRedisListConfig() RedisListConfig {
	setRedisListConfigDefault()
	return RedisListConfig{
		QueueName: config.Get[string](config.QueueName),
		BatchSize: config.Get[int64](config.QueueBatchSize),
		IsLIFO:    config.Get[bool](config.RedisEnableLIFO),
	}
}

// RedisList is the Redis implementation of queue.Service using redis lists (First-In First-Out or Last-In First-Out).
type RedisList struct {
	Client *redis.Client
	Codec  codec.Codec
	Config RedisListConfig
}

var _ Queue = RedisList{}

func NewRedisList(cfg RedisListConfig, c *redis.Client) RedisList {
	return RedisList{
		Client: c,
		Codec:  codec.JSON{},
		Config: cfg,
	}
}

func (r RedisList) Pop(ctx context.Context) ([]boltzmann.Task, error) {
	// using optimistic locking through redis WATCH command and
	// ensuring atomicity between ops by using a transaction and pipelines.
	var tasks []boltzmann.Task
	err := r.Client.Watch(ctx, func(tx *redis.Tx) error {
		cmd := tx.LRange(ctx, r.Config.QueueName, 0, r.Config.BatchSize-1)
		if err := cmd.Err(); err != nil {
			return err
		}

		tasksEncoded, err := cmd.Result()
		if err != nil {
			return err
		}

		tasks = make([]boltzmann.Task, 0, len(tasksEncoded))
		for _, taskEncoded := range tasksEncoded {
			task := boltzmann.Task{}
			if err = r.Codec.Decode([]byte(taskEncoded), &task); err != nil {
				return err
			}

			tasks = append(tasks, task)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			return pipe.LTrim(ctx, r.Config.QueueName, int64(len(tasks))+1, -1).Err()
		})
		return err
	})

	return tasks, err
}

func (r RedisList) Push(ctx context.Context, tasks ...boltzmann.Task) error {
	encodedTasks := make([]any, 0, len(tasks))
	for _, task := range tasks {
		encodedTask, err := r.Codec.Encode(task)
		if err != nil {
			return err
		}
		encodedTasks = append(encodedTasks, encodedTask)
	}

	if r.Config.IsLIFO {
		return r.Client.LPush(ctx, r.Config.QueueName, encodedTasks...).Err()
	}

	return r.Client.RPush(ctx, r.Config.QueueName, encodedTasks...).Err()
}
