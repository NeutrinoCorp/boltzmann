package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredisync "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/execplan"
	"github.com/neutrinocorp/boltzmann/internal/codec"
	"github.com/neutrinocorp/boltzmann/internal/concurrency/lock"
	"github.com/neutrinocorp/boltzmann/internal/executor"
	"github.com/neutrinocorp/boltzmann/internal/executor/delegate"
	"github.com/neutrinocorp/boltzmann/internal/id"
	"github.com/neutrinocorp/boltzmann/internal/polling"
	"github.com/neutrinocorp/boltzmann/internal/queue"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/state"
	"github.com/neutrinocorp/boltzmann/task"
)

func main() {
	redisOpts := &redis.UniversalOptions{
		Addrs:                 []string{":6379"},
		ClientName:            fmt.Sprintf("boltzmann-node-%d", rand.Int()),
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

	redsyncPool := goredisync.NewPool(rdb)
	redsyncClient := redsync.New(redsyncPool)
	distLockCfg := lock.DistributedLockConfig{
		LeaseDuration: time.Second * 15,
	}
	repoConfig := boltzmann.RepositoryConfig{
		ItemTTL: time.Hour,
	}
	stateRepo := state.RepositoryRedis{
		Config: repoConfig,
		Codec:  codec.Msgpack{},
		RDB:    rdb,
	}
	stateSvc := state.Service[boltzmann.Task]{
		Repository: stateRepo,
	}
	stateExecPlanSvc := state.Service[execplan.ExecutionPlanReference]{
		Repository: stateRepo,
	}
	agentExecDelegate := task.Delegate{
		Service: task.Service{
			AgentRegistry: agent.Registry{},
		},
	}
	var delegateExecutor delegate.Delegate[boltzmann.Task] = delegate.LockingMiddleware[string, boltzmann.Task]{
		Lock: lock.NewRedisLock[boltzmann.Task](distLockCfg, redsyncClient),
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

	execPlanQueue := queue.Redis[execplan.ExecutionPlanReference]{
		QueueConfig: queue.Config{
			QueueName: "boltzmann-exec-plan-queue",
		},
		Client: rdb,
	}
	execPlanRepo := execplan.RepositoryRedis{
		Config: repoConfig,
		RDB:    rdb,
		Codec:  codec.Msgpack{},
	}
	execPlanSvc := execplan.Service{
		Repository: execPlanRepo,
		TaskExecutorDelegate: executor.SyncExecutor[boltzmann.Task]{
			Delegate: agentExecDelegate,
		},
	}
	svc := scheduler.Service{
		Config: scheduler.ServiceConfig{
			MaxScheduledTasks:    512,
			PayloadTruncateLimit: 512,
		},
		TaskQueue:       workQueue,
		ExecPlanQueue:   execPlanQueue,
		ExecPlanService: execPlanSvc,
		CodecStrategy:   codec.NewStrategy(),
		FactoryID:       id.NewKSUID,
		AgentRegistry:   agent.Registry{},
	}
	ctrl := scheduler.ControllerHTTP{
		Service: svc,
	}
	taskCtrl := task.ControllerHTTP{Service: stateSvc}
	execPlanCtrl := execplan.ControllerHTTP{
		Service:      execPlanSvc,
		StateService: stateExecPlanSvc,
	}

	e := echo.New()
	ctrl.SetRoutes(e)
	taskCtrl.SetRoutes(e)
	execPlanCtrl.SetRoutes(e)
	e.Debug = true

	var delegateExecPlanExecutor delegate.Delegate[execplan.ExecutionPlanReference] = delegate.LockingMiddleware[string, execplan.ExecutionPlanReference]{
		Lock: lock.NewRedisLock[execplan.ExecutionPlanReference](distLockCfg, redsyncClient),
		Next: delegate.CommitterMiddleware[string, execplan.ExecutionPlanReference]{
			StateService: stateExecPlanSvc,
			Next: execplan.Delegate{
				Service: execPlanSvc,
			},
		},
	}
	pollingSvcExecPlan := polling.Service[execplan.ExecutionPlanReference]{
		Config: polling.Config{
			Name:             "exec-plan-poller",
			PollInterval:     time.Second * 3,
			RetryInterval:    time.Second * 5,
			MaxRetries:       5,
			BatchSizePerPoll: 100,
		},
		Queue: execPlanQueue,
		ExecutorService: &executor.ConcurrentExecutor[execplan.ExecutionPlanReference]{
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
