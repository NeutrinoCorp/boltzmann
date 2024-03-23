package scheduler

import (
	"github.com/hashicorp/go-multierror"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/execplan"
)

// Converts a ScheduleTasksCommand into an execplan.ExecutionPlan.
func convertCommandToExecPlan(s *Service, cmd ScheduleTasksCommand) (execplan.ExecutionPlan, error) {
	planID, err := s.FactoryID()
	if err != nil {
		return execplan.ExecutionPlan{}, err
	}

	execPlan := execplan.ExecutionPlan{
		PlanID:       planID,
		WithFairness: cmd.WithFairness,
	}
	tasks := make([]boltzmann.Task, 0, len(cmd.Tasks))
	errs := &multierror.Error{}

	for i, taskCmd := range cmd.Tasks {
		if errEx := s.AgentRegistry.Exists(taskCmd.Driver); errEx != nil {
			errs = multierror.Append(err, errEx)
			continue
		}

		currentTask := boltzmann.Task{
			TaskID:          i,
			ExecutionPlanID: planID,
			Driver:          taskCmd.Driver,
			ResourceURL:     taskCmd.ResourceURL,
			AgentArguments:  taskCmd.AgentArguments,
			TypeMIME:        taskCmd.TypeMIME,
			Payload:         taskCmd.Payload,
		}
		if currentTask.TypeMIME == "" {
			tasks = append(tasks, currentTask)
			continue
		}

		currentTask.EncodedPayload, err = s.CodecStrategy.EncodeWithTruncation(currentTask.TypeMIME, s.Config.PayloadTruncateLimit, currentTask.Payload)
		if err != nil {
			errs = multierror.Append(err, err)
			continue
		}

		tasks = append(tasks, currentTask)
	}
	if errs.Len() > 0 {
		return execplan.ExecutionPlan{}, errs
	}

	execPlan.Tasks = tasks
	return execPlan, nil
}
