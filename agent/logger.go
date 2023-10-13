package agent

import (
	"context"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
)

var logger = log.With().Str("component", "agent.logger").Logger()

type Logger struct {
	next Agent
}

var _ Middleware = &Logger{}

func (l *Logger) SetNext(a Agent) {
	l.next = a
}

func (l *Logger) Execute(ctx context.Context, task boltzmann.Task) (io.ReadCloser, error) {
	logger.Info().
		Str("task_id", task.TaskID).
		Str("correlation_id", task.CorrelationID).
		Str("driver", task.Driver).
		Str("resource_location", task.ResourceURI).
		Msg("executing agent task")

	res, err := l.next.Execute(ctx, task)
	if err != nil {
		logger.Err(err).
			Str("task_id", task.TaskID).
			Str("correlation_id", task.CorrelationID).
			Str("driver", task.Driver).
			Str("resource_location", task.ResourceURI).
			Msg("failed to execute agent task")
		return nil, err
	}

	logger.Info().
		Str("task_id", task.TaskID).
		Str("correlation_id", task.CorrelationID).
		Str("driver", task.Driver).
		Str("resource_location", task.ResourceURI).
		Msg("successfully executed agent task")
	return res, nil
}
