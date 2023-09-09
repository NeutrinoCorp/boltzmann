package queue

import "github.com/rs/zerolog/log"

var (
	embeddedSvcLogger = log.With().Str("component", embeddedServiceModule).Logger()
	redisSvcLogger    = log.With().Str("component", redisServiceModule).Logger()
)
