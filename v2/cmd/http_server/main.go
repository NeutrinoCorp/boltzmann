package main

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredisync "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/agent"
	"github.com/neutrinocorp/boltzmann/v2/codec"
	"github.com/neutrinocorp/boltzmann/v2/concurrency/lock"
	"github.com/neutrinocorp/boltzmann/v2/execplan"
	"github.com/neutrinocorp/boltzmann/v2/executor"
	"github.com/neutrinocorp/boltzmann/v2/executor/delegate"
	"github.com/neutrinocorp/boltzmann/v2/id"
	"github.com/neutrinocorp/boltzmann/v2/polling"
	"github.com/neutrinocorp/boltzmann/v2/queue"
	"github.com/neutrinocorp/boltzmann/v2/scheduler"
	"github.com/neutrinocorp/boltzmann/v2/state"
	"github.com/neutrinocorp/boltzmann/v2/task"
)

func main() {
	redisOpts := &redis.UniversalOptions{
		Addrs:                 []string{":6379"},
		ClientName:            "boltzmann-node",
		DB:                    0,
		Dialer:                nil,
		OnConnect:             nil,
		Protocol:              0,
		Username:              "",
		Password:              "",
		SentinelUsername:      "",
		SentinelPassword:      "",
		MaxRetries:            0,
		MinRetryBackoff:       0,
		MaxRetryBackoff:       0,
		DialTimeout:           0,
		ReadTimeout:           0,
		WriteTimeout:          0,
		ContextTimeoutEnabled: false,
		PoolFIFO:              false,
		PoolSize:              0,
		PoolTimeout:           0,
		MinIdleConns:          0,
		MaxIdleConns:          0,
		ConnMaxIdleTime:       0,
		ConnMaxLifetime:       0,
		TLSConfig:             nil,
		MaxRedirects:          0,
		ReadOnly:              false,
		RouteByLatency:        false,
		RouteRandomly:         false,
		MasterName:            "",
	}
	rdb := redis.NewUniversalClient(redisOpts)
	defer rdb.Close()

	// TODO: Add validator with go-playground/validator/v10. For both ctrl request payloads and configs.
	redsyncPool := goredisync.NewPool(rdb)
	redsyncClient := redsync.New(redsyncPool)
	lockFactoryDelegate := lock.RedlockFactory{
		RedsyncClient: redsyncClient,
		Config: lock.DistributedLockConfig{
			LeaseExpireDuration: time.Second * 5,
		},
	}

	repoConfig := boltzmann.RepositoryConfig{
		ItemTTL: time.Hour,
	}
	stateRepo := state.RepositoryRedis{
		Config: repoConfig,
		Codec:  codec.JSON{},
		RDB:    rdb,
	}
	stateSvc := state.Service[boltzmann.Task]{
		Repository: stateRepo,
	}
	stateExecPlanSvc := state.Service[boltzmann.ExecutionPlanReference]{
		Repository: stateRepo,
	}
	agentExecDelegate := agent.ExecutorDelegate{}
	var delegateExecutor delegate.Delegate[boltzmann.Task] = delegate.LockingMiddleware[string, boltzmann.Task]{
		LockFactory: lockFactoryDelegate,
		Next: delegate.CommitterMiddleware[string, boltzmann.Task]{
			StateService: stateSvc,
			Next:         agentExecDelegate,
		},
	}
	workQueue := queue.Redis[boltzmann.Task]{
		QueueConfig: queue.Config{
			QueueName: "boltzmann-job-queue",
		},
		Client: rdb,
	}
	pollingSvc := polling.Service[boltzmann.Task]{
		Config: polling.Config{
			Name:             "task-poller",
			PollInterval:     time.Second * 3,
			RetryInterval:    time.Second * 5,
			MaxRetries:       5,
			BatchSizePerPoll: 100,
		},
		Queue: workQueue,
		ExecutorService: &executor.ConcurrentExecutor[boltzmann.Task]{
			Config: executor.ConcurrentExecutorConfig{
				MaxGoroutines: 10,
			},
			Delegate: delegateExecutor,
		},
	}

	execPlanQueue := queue.Redis[boltzmann.ExecutionPlanReference]{
		QueueConfig: queue.Config{
			QueueName: "boltzmann-exec-plan-queue",
		},
		Client: rdb,
	}
	execPlanRepo := execplan.RepositoryRedis{
		Config: repoConfig,
		RDB:    rdb,
		Codec:  codec.JSON{},
	}
	execPlanSvc := execplan.Service{
		Repository: execPlanRepo,
		Queue:      execPlanQueue,
		TaskExecutorDelegate: executor.SyncExecutor[boltzmann.Task]{
			Delegate: agentExecDelegate,
		},
	}
	svc := scheduler.Service{
		Config: scheduler.ServiceConfig{
			PayloadTruncateLimit: 512,
		},
		TaskQueue:       workQueue,
		ExecPlanService: execPlanSvc,
		CodecStrategy:   codec.Strategy{},
		FactoryID:       id.NewKSUID,
		AgentRegistry:   agent.Registry{},
	}
	ctrl := scheduler.ControllerHTTP{
		Service: svc,
	}
	taskCtrl := task.ControllerHTTP{Service: stateSvc}
	execPlanCtrl := execplan.ControllerHTTP{Service: stateExecPlanSvc}

	e := echo.New()
	ctrl.SetRoutes(e)
	taskCtrl.SetRoutes(e)
	execPlanCtrl.SetRoutes(e)
	e.Debug = true

	var delegateExecPlanExecutor delegate.Delegate[boltzmann.ExecutionPlanReference] = delegate.LockingMiddleware[string, boltzmann.ExecutionPlanReference]{
		LockFactory: lockFactoryDelegate,
		Next: delegate.CommitterMiddleware[string, boltzmann.ExecutionPlanReference]{
			StateService: stateExecPlanSvc,
			Next: execplan.Delegate{
				Service: execPlanSvc,
			},
		},
	}
	pollingSvcExecPlan := polling.Service[boltzmann.ExecutionPlanReference]{
		Config: polling.Config{
			Name:             "exec-plan-poller",
			PollInterval:     time.Second * 3,
			RetryInterval:    time.Second * 5,
			MaxRetries:       5,
			BatchSizePerPoll: 100,
		},
		Queue: execPlanQueue,
		ExecutorService: &executor.ConcurrentExecutor[boltzmann.ExecutionPlanReference]{
			Config: executor.ConcurrentExecutorConfig{
				MaxGoroutines: 10,
			},
			Delegate: delegateExecPlanExecutor,
		},
	}
	rootCtx, rootCtxCancelFunc := context.WithCancel(context.Background())
	go func() {
		if err := pollingSvc.Start(rootCtx); err != nil {
			log.Err(err).Str("poller_name", "task").Msg("stopping polling, got error")
		}
	}()
	go func() {
		if err := pollingSvcExecPlan.Start(rootCtx); err != nil {
			log.Err(err).Str("poller_name", "exec_plan").Msg("stopping polling, got error")
		}
	}()

	if err := e.Start(":8080"); err != nil {
		rootCtxCancelFunc()
		panic(err)
	}
	rootCtxCancelFunc()
}
