package main

import (
	"context"
	"net/http"
	"runtime"

	"github.com/neutrinocorp/boltzmann/codec"

	"github.com/labstack/echo/v4"
	"github.com/neutrinocorp/boltzmann/controller"

	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/queue"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/state"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Register agent drivers
	agentReg := agent.Registry{}
	agentReg[agent.HTTPDriverName] = agent.HTTP{
		Client: http.DefaultClient,
	}

	// 2. Setup state storage
	redisCfg := &redis.Options{
		Network:               "",
		Addr:                  "localhost:6379",
		ClientName:            "",
		Dialer:                nil,
		OnConnect:             nil,
		Protocol:              0,
		Username:              "",
		Password:              "",
		CredentialsProvider:   nil,
		DB:                    0,
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
		Limiter:               nil,
	}
	redisClient := redis.NewClient(redisCfg)
	stateStore := state.RedisRepository{
		Client: redisClient,
		Codec:  codec.JSON{},
	}

	// 3. Setup queueing service
	queueSvcCfg := queue.EmbeddedServiceConfig{
		BufferSize: 100,
		MaxWorkers: int64(runtime.GOMAXPROCS(0)),
	}
	queueSvc := queue.NewEmbeddedService(queueSvcCfg, agentReg, stateStore)

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
	go queueSvc.Start(context.Background())

	// 6b. Setup and start REST HTTP server
	e := echo.New()
	ctrl := controller.TaskSchedulerHTTP{
		Service: svc,
	}
	versionedRouter := e.Group("/api/v1")
	ctrl.SetRoutes(versionedRouter)

	if err := e.Start(":8081"); err != nil {
		panic(err)
	}

	// 7. Shutdown background services (queueing, supervisor)
	err := queueSvc.Shutdown()
	if err != nil {
		panic(err)
	}
}
