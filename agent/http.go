package agent

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
)

const (
	HTTPDriverName = "http"

	HTTPMethodArgKey = "agent.method"
)

type HTTP struct {
	Client *http.Client
}

var _ Agent = HTTP{}

func (h HTTP) Execute(ctx context.Context, task boltzmann.Task) (io.ReadCloser, error) {
	log.Info().Msg("executing http task")

	method := task.AgentArguments[HTTPMethodArgKey]
	var reader io.Reader
	if len(task.Payload) > 0 && (method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut) {
		reader = bytes.NewReader(task.Payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, task.ResourceURI, reader)
	if err != nil {
		return nil, err
	}

	for k, v := range task.AgentArguments {
		if strings.HasPrefix(k, "agent.") {
			continue
		}

		req.Header.Add(k, v)
	}

	logger.Info().
		Str("task_id", task.TaskID).
		Str("driver", task.Driver).
		Str("resource_location", task.ResourceURI).
		Str("method", method).
		Int64("request_content_length", req.ContentLength).
		Int("request_header_total_entries", len(req.Header)).
		Msg("sending http request")
	res, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}

	logger.Info().
		Str("task_id", task.TaskID).
		Str("driver", task.Driver).
		Str("resource_location", task.ResourceURI).
		Str("method", method).
		Int("status_code", res.StatusCode).
		Int64("content_length", res.ContentLength).
		Msg("got http response")
	return res.Body, nil
}
