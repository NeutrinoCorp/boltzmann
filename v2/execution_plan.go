package boltzmann

type ExecutionPlan struct {
	PlanID       string
	Tasks        []Task
	WithFairness bool
}

var _ Identifiable[string] = ExecutionPlan{}

func (e ExecutionPlan) GetID() string {
	return e.PlanID
}
