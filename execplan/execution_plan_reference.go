package execplan

import "github.com/neutrinocorp/boltzmann"

type ExecutionPlanReference struct {
	PlanID string
}

var _ boltzmann.Identifiable[string] = ExecutionPlanReference{}

func (e ExecutionPlanReference) GetID() string {
	return e.PlanID
}
