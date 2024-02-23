package state

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/codec"
)

type Repository interface {
	Save(ctx context.Context, item State) error
	GetByID(ctx context.Context, key string) (State, error)
}

type RepositoryRedis struct {
	Config boltzmann.RepositoryConfig
	Codec  codec.Codec
	RDB    redis.UniversalClient
}

var _ Repository = RepositoryRedis{}

func (r RepositoryRedis) Save(ctx context.Context, item State) error {
	encodedItem, err := r.Codec.Encode(item)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("states#%s", item.ID)
	return r.RDB.Set(ctx, key, encodedItem, r.Config.ItemTTL).Err()
}

func (r RepositoryRedis) GetByID(ctx context.Context, key string) (State, error) {
	redisKey := fmt.Sprintf("states#%s", key)
	encodedItem, err := r.RDB.Get(ctx, redisKey).Bytes()
	if err != nil && redis.HasErrorPrefix(err, "redis: nil") {
		return State{}, boltzmann.ErrItemNotFound{
			ResourceName: "state",
			ResourceID:   redisKey,
		}
	} else if err != nil {
		return State{}, err
	}

	item := State{}
	if err = r.Codec.Decode(encodedItem, &item); err != nil {
		return State{}, err
	}

	return item, nil
}
