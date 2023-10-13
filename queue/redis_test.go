//go:build !integration

package queue_test

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
)

type redisListTestSuite struct {
	suite.Suite

	client *redis.Client
}

func TestRedisList(t *testing.T) {
	suite.Run(t, &redisListTestSuite{})
}

func (r *redisListTestSuite) SetupSuite() {
	redisUrl := "redis://@localhost:6379/0?dial_timeout=3&read_timeout=6s&max_retries=2"
	redisCfg, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Err(err).Msg("cannot read redis url")
		return
	}
	r.client = redis.NewClient(redisCfg)
}

func (r *redisListTestSuite) TearDownSuite() {
	_ = r.client.Close()
}
