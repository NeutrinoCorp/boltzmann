package queue

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann/internal/codec"
)

type Redis[T any] struct {
	QueueConfig Config
	Client      redis.UniversalClient
	Codec       codec.Msgpack
}

var _ Queue[string] = Redis[string]{}

func (r Redis[T]) Push(ctx context.Context, items ...T) error {
	args := make([]any, 0, len(items))
	for _, item := range items {
		itemJSON, err := r.Codec.Encode(item)
		if err != nil {
			return err
		}
		args = append(args, itemJSON)
	}

	return r.Client.LPush(ctx, r.QueueConfig.QueueName, args...).Err()
}

func (r Redis[T]) Pop(ctx context.Context, popRange int) ([]T, error) {
	resultsRaw, err := r.Client.LPopCount(ctx, r.QueueConfig.QueueName, popRange).Result()
	if err != nil && !redis.HasErrorPrefix(err, "redis: nil") {
		return nil, err
	}

	results := make([]T, 0, len(resultsRaw))
	errs := &multierror.Error{}
	for _, resRaw := range resultsRaw {
		var res T
		if errDecode := r.Codec.Decode([]byte(resRaw), &res); errDecode != nil {
			errs = multierror.Append(errs, errDecode)
			continue
		}
		results = append(results, res)
	}
	if errs.Len() > 0 {
		return nil, errs
	}

	return results, nil
}
