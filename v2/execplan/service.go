package execplan

import (
	"context"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/executor"
	"github.com/neutrinocorp/boltzmann/v2/queue"
)

type Service struct {
	Repository           Repository
	Queue                queue.Queue[boltzmann.ExecutionPlanReference]
	TaskExecutorDelegate executor.SyncExecutor[boltzmann.Task]
}

func (s Service) CreateAndEnqueue(ctx context.Context, plan boltzmann.ExecutionPlan) error {
	if err := s.Repository.Save(ctx, plan); err != nil {
		return err
	}

	return s.Queue.Push(ctx, boltzmann.ExecutionPlanReference{PlanID: plan.PlanID})
}

func (s Service) RunPlan(ctx context.Context, planRef boltzmann.ExecutionPlanReference) error {
	plan, err := s.Repository.GetByID(ctx, planRef.PlanID)
	if err != nil {
		return err
	}

	return s.TaskExecutorDelegate.ExecuteAll(ctx, plan.Tasks)
}
