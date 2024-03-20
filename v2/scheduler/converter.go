package scheduler

import (
	"bytes"
	"io"

	"github.com/hashicorp/go-multierror"

	"github.com/neutrinocorp/boltzmann/v2"
)

func convertCommandToExecPlan(s *Service, cmd ScheduleTasksCommand) (boltzmann.ExecutionPlan, error) {
	planID, err := s.FactoryID()
	if err != nil {
		return boltzmann.ExecutionPlan{}, err
	}

	execPlan := boltzmann.ExecutionPlan{
		PlanID:       planID,
		WithFairness: cmd.WithFairness,
	}
	tasks := make([]boltzmann.Task, 0, len(cmd.Tasks))
	errs := &multierror.Error{}

	payloadWriter := bytes.NewBuffer(nil)
	for i, taskCmd := range cmd.Tasks {
		if errEx := s.AgentRegistry.Exists(taskCmd.Driver); errEx != nil {
			errs = multierror.Append(err, errEx)
			continue
		}

		task := boltzmann.Task{
			TaskID:          i,
			ExecutionPlanID: planID,
			Driver:          taskCmd.Driver,
			ResourceURL:     taskCmd.ResourceURL,
			AgentArguments:  taskCmd.AgentArguments,
			TypeMIME:        taskCmd.TypeMIME,
			Payload:         taskCmd.Payload,
		}
		if task.TypeMIME == "" {
			tasks = append(tasks, task)
			continue
		}

		encodedPayload, errEncode := s.CodecStrategy.Encode(task.TypeMIME, task.Payload)
		if errEncode != nil {
			errs = multierror.Append(err, errEncode)
			continue
		}

		payloadWriter.Reset()
		reader := bytes.NewReader(encodedPayload)
		lr := io.LimitReader(reader, s.Config.PayloadTruncateLimit)
		if _, err = payloadWriter.ReadFrom(lr); err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		task.EncodedPayload = payloadWriter.Bytes()
		tasks = append(tasks, task)
	}
	if errs.Len() > 0 {
		return boltzmann.ExecutionPlan{}, errs
	}

	execPlan.Tasks = tasks
	return execPlan, nil
}
