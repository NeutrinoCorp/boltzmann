package execplan

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/internal/codec"
)

type Repository interface {
	Save(ctx context.Context, plan ExecutionPlan) error
	GetByID(ctx context.Context, planID string) (ExecutionPlan, error)
}

type RepositoryRedis struct {
	Config boltzmann.RepositoryConfig
	RDB    redis.UniversalClient
	Codec  codec.Msgpack
}

var _ Repository = RepositoryRedis{}

func (r RepositoryRedis) Save(ctx context.Context, plan ExecutionPlan) error {
	encodedPlan, err := r.Codec.Encode(plan)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("exec_plan#%s", plan.PlanID)
	return r.RDB.Set(ctx, key, encodedPlan, r.Config.ItemTTL).Err()
}

func (r RepositoryRedis) GetByID(ctx context.Context, planID string) (ExecutionPlan, error) {
	key := fmt.Sprintf("exec_plan#%s", planID)
	encodedPlan, err := r.RDB.Get(ctx, key).Bytes()
	if err != nil && redis.HasErrorPrefix(err, "nil") {
		return ExecutionPlan{}, errors.New("entity not found")
	}
	if err != nil {
		return ExecutionPlan{}, err
	}

	plan := ExecutionPlan{}
	err = r.Codec.Decode(encodedPlan, &plan)
	return plan, err
}
