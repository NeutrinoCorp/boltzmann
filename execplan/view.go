package execplan

import "github.com/neutrinocorp/boltzmann/task"

type View struct {
	PlanID       string      `json:"plan_id"`
	WithFairness bool        `json:"with_fairness"`
	Tasks        []task.View `json:"tasks"`
}
