package queue

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann/v2/codec"
)

type Redis[T any] struct {
	Client redis.UniversalClient
	Codec  codec.JSON // redis ONLY accepts JSON codec
}

var _ Queue[string] = Redis[string]{}

func (r Redis[T]) Push(ctx context.Context, queueName string, items []T) error {
	args := make([]any, 0, len(items))
	for _, item := range items {
		itemJSON, err := r.Codec.Encode(item)
		if err != nil {
			return err
		}
		args = append(args, itemJSON)
	}

	// TODO: Use PubSub for no FIFO requirement (no fairness)
	return r.Client.RPush(ctx, queueName, args...).Err()
}
