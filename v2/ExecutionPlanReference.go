package boltzmann

type ExecutionPlanReference struct {
	PlanID string
}

var _ Identifiable[string] = ExecutionPlanReference{}

func (e ExecutionPlanReference) GetID() string {
	return e.PlanID
}
