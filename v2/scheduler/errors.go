package scheduler

import "github.com/pkg/errors"

var (
	ErrNoExecutionPlan = errors.New("scheduler: no execution plan was defined")
)
