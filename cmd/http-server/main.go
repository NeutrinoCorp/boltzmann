package main

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/codec"
	"github.com/neutrinocorp/boltzmann/controller"
	"github.com/neutrinocorp/boltzmann/queue"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/state"
)

func main() {
	// 1. Register agent drivers
	agentReg := agent.Registry{}
	agentReg.Register(agent.HTTPDriverName, agent.HTTP{
		Client: http.DefaultClient,
	})

	// 2. Setup state storage
	redisCfg := &redis.Options{
		Addr: "localhost:6379",
	}
	redisClient := redis.NewClient(redisCfg)
	stateStore := state.RedisRepository{
		Client: redisClient,
		Codec:  codec.JSON{},
	}

	// 3. Setup queueing service
	// queueSvcCfg := queue.EmbeddedServiceConfig{
	// 	BufferSize:           100,
	// 	MaxInFlightProcesses: int64(runtime.GOMAXPROCS(0)),
	// }
	// queueSvc := queue.NewEmbeddedService(queueSvcCfg, agentReg, stateStore)
	queueSvcCfg := queue.RedisServiceConfig{
		StreamName:                    "boltzmann-job-queue",
		StreamGroupID:                 "boltzmann-agent-worker_pool",
		MaxInFlightProcesses:          int64(runtime.GOMAXPROCS(0)),
		EnableStreamGroupAutoCreation: true,
		RetryBackoff:                  time.Second * 5,
	}
	queueSvc := queue.NewRedisService(redisClient, queueSvcCfg)

	// 4. Setup task scheduler
	sched := scheduler.SyncTaskScheduler{
		AgentRegistry:   agentReg,
		QueueService:    queueSvc,
		StateRepository: stateStore,
	}

	// 5. Setup service
	svc := scheduler.Service{
		Scheduler:       sched,
		StateRepository: stateStore,
	}

	// 6. Start internal background services (queueing, supervisor).
	go func() {
		err := queueSvc.Start(context.Background())
		if err != nil {
			log.Err(err).Msg("failed to start queue service")
		}
	}()

	// 7. Setup and start REST HTTP server
	e := echo.New()
	ctrl := controller.TaskSchedulerHTTP{
		Service: svc,
	}
	versionedRouter := e.Group("/api/v1")
	ctrl.SetRoutes(versionedRouter)

	if err := e.Start(":8081"); err != nil {
		panic(err)
	}

	// 8. Shutdown background services (queueing, supervisor)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	err := queueSvc.Shutdown(shutdownCtx)
	if err != nil {
		panic(err)
	}
}
