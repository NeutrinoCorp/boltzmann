package queue

import "github.com/rs/zerolog/log"

var internalSvcLogger = log.With().Str("component", internalServiceModule).Logger()
