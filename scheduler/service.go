package scheduler

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/execplan"
	"github.com/neutrinocorp/boltzmann/internal/codec"
	"github.com/neutrinocorp/boltzmann/internal/id"
	"github.com/neutrinocorp/boltzmann/internal/queue"
)

// Service is a service component used by external systems to schedule tasks.
type Service struct {
	Config          ServiceConfig
	TaskQueue       queue.Queue[boltzmann.Task]
	ExecPlanQueue   queue.Queue[execplan.ExecutionPlanReference]
	ExecPlanService execplan.Service
	CodecStrategy   codec.Strategy
	FactoryID       id.FactoryFunc
	AgentRegistry   agent.Registry
}

// Schedule schedules one or many tasks, creates a boltzmann.ExecutionPlan.
//
// According to ScheduleTasksCommand.WithFairness flag, the routine will use different scheduling mechanisms.
//
// If FALSE, each task requested will be sent to a task-only queue to be later executed concurrently.
//
// If TRUE,
// an execution plan will be stored and then enqueued into an execution plan-only queue
// to later execute tasks sequentially.
// Moreover, a `claim-check` messaging pattern is used as an execution plan might contain many tasks.
func (s Service) Schedule(ctx context.Context, cmd ScheduleTasksCommand) (execplan.ExecutionPlan, error) {
	if err := boltzmann.GlobalValidator.StructCtx(ctx, cmd); err != nil {
		return execplan.ExecutionPlan{}, err
	} else if len(cmd.Tasks) > s.Config.MaxScheduledTasks {
		return execplan.ExecutionPlan{}, boltzmann.NewOutOfRangeWithType[execplan.ExecutionPlan, int]("tasks",
			1, s.Config.MaxScheduledTasks)
	}

	execPlan, err := convertCommandToExecPlan(&s, cmd)
	if err != nil {
		return execplan.ExecutionPlan{}, err
	}

	if err = s.ExecPlanService.Create(ctx, execPlan); err != nil {
		return execplan.ExecutionPlan{}, err
	}

	if !cmd.WithFairness {
		return execPlan, s.TaskQueue.Push(ctx, execPlan.Tasks...)
	}
	return execPlan, s.ExecPlanQueue.Push(ctx, execplan.ExecutionPlanReference{PlanID: execPlan.PlanID})
}
