package execplan

import (
	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/task"
)

// ExecutionPlan aggregate with a set of tasks to be executed by a system. Contains a `WithConcurrency` flag to
// indicate the usage of
type ExecutionPlan struct {
	PlanID       string           // Execution plan unique identifier.
	Tasks        []boltzmann.Task // Set of tasks to be executed.
	WithFairness bool             // Indicates whether to respect ordering of the given set of tasks or not.
}

var _ boltzmann.Identifiable[string] = ExecutionPlan{}

func (e ExecutionPlan) GetID() string {
	return e.PlanID
}

func (e ExecutionPlan) View() View {
	tasks := make([]task.View, 0, len(e.Tasks))
	for _, item := range e.Tasks {
		tasks = append(tasks, task.View{
			TaskID:          item.TaskID,
			ExecutionPlanID: item.ExecutionPlanID,
			Driver:          item.Driver,
			ResourceURL:     item.ResourceURL,
			AgentArguments:  item.AgentArguments,
			TypeMIME:        item.TypeMIME,
			PayloadSize:     len(item.EncodedPayload),
		})
	}

	return View{
		PlanID:       e.PlanID,
		WithFairness: e.WithFairness,
		Tasks:        tasks,
	}
}
