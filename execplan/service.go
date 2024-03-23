package execplan

import (
	"context"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/internal/executor"
)

type Service struct {
	Repository           Repository
	TaskExecutorDelegate executor.SyncExecutor[boltzmann.Task]
}

func (s Service) Create(ctx context.Context, plan ExecutionPlan) error {
	return s.Repository.Save(ctx, plan)
}

func (s Service) RunPlan(ctx context.Context, planRef ExecutionPlanReference) error {
	plan, err := s.Repository.GetByID(ctx, planRef.PlanID)
	if err != nil {
		return err
	}

	return s.TaskExecutorDelegate.ExecuteAll(ctx, plan.Tasks)
}

func (s Service) FindByID(ctx context.Context, id string) (ExecutionPlan, error) {
	return s.Repository.GetByID(ctx, id)
}
