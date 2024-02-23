package execplan

import (
	"context"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/executor/delegate"
)

type Delegate struct {
	Service Service
}

var _ delegate.Delegate[boltzmann.ExecutionPlanReference] = Delegate{}

func (e Delegate) Execute(ctx context.Context, planRef boltzmann.ExecutionPlanReference) error {
	return e.Service.RunPlan(ctx, planRef)
}
