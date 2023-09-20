package agent

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
)

const (
	HTTPDriverName = "http"

	HTTPMethodArgKey = "http.method"
)

type HTTP struct {
	Client *http.Client
}

var _ Agent = HTTP{}

func (h HTTP) Execute(ctx context.Context, task boltzmann.Task) error {
	log.Info().Msg("executing http task")
	req, err := http.NewRequestWithContext(ctx, task.AgentArguments[HTTPMethodArgKey],
		task.ResourceURI, nil)
	if err != nil {
		return err
	}

	res, err := h.Client.Do(req)
	if err != nil {
		return err
	}
	log.Info().
		Str("task_id", task.TaskID).
		Str("driver", task.Driver).
		Str("resource_location", task.ResourceURI).
		Int("status_code", res.StatusCode).
		Int64("content_length", res.ContentLength).
		Msg("got http response")

	return nil
}
