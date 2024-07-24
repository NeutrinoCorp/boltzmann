package agent

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
)

type Noop struct{}

var _ Agent = Noop{}

func (n Noop) ExecTask(_ context.Context, cmd boltzmann.Task) error {
	log.Debug().Interface("task", cmd).Msg("got task at noop")
	return nil
}
