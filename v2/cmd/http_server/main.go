package main

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/neutrinocorp/boltzmann/v2/queue"
	"github.com/neutrinocorp/boltzmann/v2/scheduler"
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

	q := queue.Redis[scheduler.TaskCommand]{
		Client: rdb,
	}
	svc := scheduler.Service{
		Queue: q,
		Config: scheduler.ServiceConfig{
			QueueName: "boltzmann-job-queue-fifo",
		},
	}
	ctrl := scheduler.ControllerHTTP{
		Service: svc,
	}

	e := echo.New()
	ctrl.SetRoutes(e)

	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}
