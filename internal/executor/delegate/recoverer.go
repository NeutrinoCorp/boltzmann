package delegate

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
)

type Recoverer[I comparable, T boltzmann.Identifiable[I]] struct {
}

var _ Delegate[boltzmann.NoopIdentifiable] = Recoverer[string, boltzmann.NoopIdentifiable]{}

func (r Recoverer[I, T]) Execute(ctx context.Context, item T) error {
	//TODO implement me
	panic("implement me")
}
