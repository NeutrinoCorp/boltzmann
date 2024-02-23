package scheduler

import (
	"context"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/agent"
	"github.com/neutrinocorp/boltzmann/v2/codec"
	"github.com/neutrinocorp/boltzmann/v2/execplan"
	"github.com/neutrinocorp/boltzmann/v2/id"
	"github.com/neutrinocorp/boltzmann/v2/queue"
)

type ServiceConfig struct {
	PayloadTruncateLimit int64
}

// Service scheduling service.
type Service struct {
	Config          ServiceConfig
	TaskQueue       queue.Queue[boltzmann.Task]
	ExecPlanService execplan.Service
	CodecStrategy   codec.Strategy
	FactoryID       id.FactoryFunc
	AgentRegistry   agent.Registry
}

func (s Service) Schedule(ctx context.Context, cmd ScheduleTasksCommand) (boltzmann.ExecutionPlan, error) {
	if len(cmd.Tasks) == 0 {
		return boltzmann.ExecutionPlan{}, ErrNoExecutionPlan
	}

	// fair: waits for every step to finish sequentially to guarantee ordering in a single thread (blocking/non-blocking options for end-user)
	// unfair: will publish all tasks at the same time in work queue.
	// TODO: For fair scheduler, implement blocking scheduler and non-blocking.
	// fair sched: For each step, wait until success. Then run following step and so on. If failure, stop execution plan.
	//  - blocking: blocks request thread (is it ACTUALLY required? Potential timeout errors as it requires to wait for the whole exec plan to finish)
	//      IDEA: Maybe websockets/gRPC-multiplex could make sense here? Add another routine to sched svc API to handle these.
	//  - non-blocking: uses an additional queue with exec plan as payload. Process in background with another polling svc.
	// IDEA:Maybe a leader-election/consensus algo will be required to coordinate schedulers handling fair exec plans
	// (in boltzmann cluster).
	//  - Maybe consistent hashing could work here
	//  - This could be reused by Supervisor instances as well
	// IDEA: Maybe is required to create an sched interface and separate fair and unfair sched.
	//  Controller would select desired by exec plan from command.

	// Use dist lock for agent worker coordination. Use raft consensus for supervisor and scheduler coordination.
	//  - Supervisor: coordinate failed task re-drive (if not overpasses max failure count threshold)
	//  - Scheduler: coordinate task execution sequence.
	execPlan, err := convertCommandToExecPlan(&s, cmd)
	if err != nil {
		return boltzmann.ExecutionPlan{}, err
	}

	if !cmd.WithFairness {
		return execPlan, s.TaskQueue.Push(ctx, execPlan.Tasks...)
	}

	return execPlan, s.ExecPlanService.CreateAndEnqueue(ctx, execPlan)
}
