package execplan

import (
	"context"

	"github.com/neutrinocorp/boltzmann/internal/executor/delegate"
)

type Delegate struct {
	Service Service
}

var _ delegate.Delegate[ExecutionPlanReference] = Delegate{}

func (e Delegate) Execute(ctx context.Context, planRef ExecutionPlanReference) error {
	return e.Service.RunPlan(ctx, planRef)
}
