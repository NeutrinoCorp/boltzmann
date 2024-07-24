package delegate

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/state"
)

// CommitterMiddleware is an implementation of Delegate that commits the state of executed process.
type CommitterMiddleware[I comparable, T boltzmann.Identifiable[I]] struct {
	StateService state.Service[T]
	Next         Delegate[T]
}

var _ Delegate[boltzmann.NoopIdentifiable] = CommitterMiddleware[string, boltzmann.NoopIdentifiable]{}

func (c CommitterMiddleware[I, T]) Execute(ctx context.Context, arg T) error {
	id := fmt.Sprintf("%v", arg.GetID())
	log.Debug().Msg("committing state")
	if err := c.StateService.Create(ctx, id); err != nil {
		return err
	}
	var err error
	defer func() {
		if err != nil {
			if errState := c.StateService.MarkAsFailed(ctx, id, err); errState != nil {
				log.Debug().Err(err).Msg("got error while committing state")
			}
			log.Debug().Err(err).Msg("committed failed state")
			return
		}

		if errState := c.StateService.MarkAsCompleted(ctx, id); errState != nil {
			log.Debug().Err(err).Msg("got error while committing state")
			return
		}
		log.Debug().Msg("committed success state")
	}()
	if err = c.Next.Execute(ctx, arg); err != nil {
		return err
	}
	return nil
}
