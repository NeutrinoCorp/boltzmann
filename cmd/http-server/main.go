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
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/agent"
	"github.com/neutrinocorp/boltzmann/codec"
	"github.com/neutrinocorp/boltzmann/config"
	"github.com/neutrinocorp/boltzmann/controller"
	"github.com/neutrinocorp/boltzmann/factory"
	"github.com/neutrinocorp/boltzmann/queue"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/state"
)

func main() {
	config.SetEnvPrefix("BOLTZMANN")
	defCodec := codec.JSON{}
	// 1. Setup state storage and service
	redisURL := config.Get[string]("REDIS_URL")
	redisCfg, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Err(err).Msg("cannot read redis url")
		return
	}
	redisClient := redis.NewClient(redisCfg)
	stateCfg := state.NewRedisRepositoryConfig()
	stateStore := state.RedisRepository{
		Client: redisClient,
		Config: stateCfg,
		Codec:  defCodec,
	}
	stateSvc := state.Service{
		StateRepository: stateStore,
		Codec:           defCodec,
	}

	// 2. Register agent drivers
	agentReg := agent.NewRegistry()
	agentReg.AddMiddleware(&agent.StateUpdater{
		Config:          agent.NewStateUpdaterConfig(),
		StateRepository: stateStore,
	})
	agentReg.AddMiddleware(&agent.Logger{})
	agentReg.AddMiddleware(&agent.Retryable{})
	agentReg.Register(agent.HTTPDriverName, agent.HTTP{
		Client: http.DefaultClient,
	})

	// 3. Setup queueing service with middlewares
	queueImpl := queue.StateUpdaterMiddleware{
		Repository: stateStore,
		Next:       queue.NewRedisList(queue.NewRedisListConfig(), defCodec, redisClient),
	}
	queueSvc := queue.NewService(queue.NewServiceConfig(), agentReg, queueImpl)

	// 4. Setup task scheduler
	sched := scheduler.TaskSchedulerDefault{
		AgentRegistry:   agentReg,
		QueueService:    queueImpl,
		StateRepository: stateStore,
	}

	// 5. Setup service
	svc := scheduler.Service{
		FactoryID:       factory.KSUID{},
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
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = controller.EchoErrHandler
	ctrl := controller.TaskSchedulerHTTP{
		Service:      svc,
		StateService: stateSvc,
	}
	versionedRouter := e.Group("/api/v1")
	ctrl.SetRoutes(versionedRouter)

	config.SetDefault("HTTP_SERVER_ADDR", ":8080")
	httpSrvAddr := config.Get[string]("HTTP_SERVER_ADDR")
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
	if err = redisClient.Close(); err != nil {
		log.Err(err).Msg("failed to stop redis client")
	}
}
