package agent

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/v2"
)

type HTTP struct {
	Client http.RoundTripper
}

var _ Agent = HTTP{}

func (a HTTP) ExecTask(ctx context.Context, cmd boltzmann.Task) error {
	payloadReader := bytes.NewReader(cmd.EncodedPayload)
	req, err := http.NewRequestWithContext(ctx, cmd.AgentArguments["http.method"], cmd.ResourceURL, payloadReader)
	if err != nil {
		return err
	}

	for k, v := range cmd.AgentArguments {
		switch k {
		case "http.method":
			req.Header.Add("Content-Type", cmd.TypeMIME)
		default:
			headerKey := strings.TrimLeft(k, "headers.")
			req.Header.Add(headerKey, v)
		}
	}
	res, err := a.Client.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Err(err).Msg("cannot decode HTTP response")
		return nil
	}

	log.Info().Bytes("body", resBody).Msg("got HTTP response")
	return nil
}
