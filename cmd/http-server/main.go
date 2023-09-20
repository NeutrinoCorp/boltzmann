package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/codec"
	"github.com/neutrinocorp/boltzmann/config"
	"github.com/neutrinocorp/boltzmann/controller"
	"github.com/neutrinocorp/boltzmann/queue"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/state"
)

func main() {
	config.DefaultEnvPrefix = "BOLTZMANN"

	// 1. Setup state storage
	config.SetDefault("REDIS_URL", "redis://@localhost:6379/0?dial_timeout=3&read_timeout=6s&max_retries=2")
	redisURL := config.GetEnv[string]("REDIS_URL")

	redisCfg, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Err(err).Msg("cannot read redis url")
		return
	}
	redisClient := redis.NewClient(redisCfg)
	stateStore := state.RedisRepository{
		Client: redisClient,
		Codec:  codec.JSON{},
	}

	// 2. Register agent drivers
	agentReg := agent.NewRegistry()
	agentReg.AddMiddleware(&agent.StateUpdater{
		StateRepository: stateStore,
	})
	agentReg.AddMiddleware(&agent.Logger{})
	agentReg.Register(agent.HTTPDriverName, agent.HTTP{
		Client: http.DefaultClient,
	})

	// 3. Setup queueing service
	queueSvcCfg := queue.NewRedisServiceConfig()
	queueSvc := queue.NewRedisService(redisClient, queueSvcCfg, agentReg, stateStore)
	// EMBEDDED
	// queueSvcCfg := queue.EmbeddedServiceConfig{
	// 	BufferSize:           100,
	// 	MaxInFlightProcesses: int64(runtime.GOMAXPROCS(0)),
	// }
	// queueSvc := queue.NewEmbeddedService(queueSvcCfg, agentReg, stateStore)

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

	// 6. Start internal background services (queueing, supervisor, server).
	go func() {
		if err = queueSvc.Start(context.Background()); err != nil {
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

	config.SetDefault("HTTP_SERVER_ADDR", ":8081")
	httpSrvAddr := config.GetEnv[string]("HTTP_SERVER_ADDR")
	go func() {
		if err = e.Start(httpSrvAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Err(err).Msg("failed to http server")
		}
	}()

	// 8. Shutdown background services (queueing, supervisor, server)
	// Wait for program closure
	shutdownSignal := make(chan os.Signal, 3)
	signal.Notify(shutdownSignal, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-shutdownSignal

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	if err = e.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("failed to stop http server")
	}
	if err = queueSvc.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("failed to stop queue service")
	}
}
