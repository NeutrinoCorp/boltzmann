package scheduler

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/queue"
	"github.com/neutrinocorp/boltzmann/state"
)

type ScheduleTaskResult struct {
	TaskID        string
	CorrelationID string
	Driver        string
	ResourceURI   string
	ScheduleTime  time.Time
}

type TaskScheduler interface {
	Schedule(ctx context.Context, tasks []boltzmann.Task) ([]ScheduleTaskResult, error)
}

type TaskSchedulerDefault struct {
	AgentRegistry   agent.Registry
	QueueService    queue.Queue
	StateRepository state.Repository
}

var _ TaskScheduler = TaskSchedulerDefault{}

func (s TaskSchedulerDefault) Schedule(ctx context.Context, tasks []boltzmann.Task) ([]ScheduleTaskResult, error) {
	results := make([]ScheduleTaskResult, 0)
	errs := &multierror.Error{}
	for i := 0; i < len(tasks); i++ {
		res := ScheduleTaskResult{
			TaskID:        tasks[i].TaskID,
			CorrelationID: tasks[i].CorrelationID,
			Driver:        tasks[i].Driver,
			ResourceURI:   tasks[i].ResourceURI,
			ScheduleTime:  tasks[i].ScheduleTime,
		}
		if _, err := s.AgentRegistry.Get(tasks[i].Driver); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		results = append(results, res)
	}

	if err := errs.ErrorOrNil(); err != nil {
		return nil, err
	}

	scopedCtx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	// push all tasks at once so concrete queue implementations can take advantage of batching APIs (if available),
	// reducing overall round-trips.
	return results, s.QueueService.Push(scopedCtx, tasks...)
}
