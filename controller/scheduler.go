package controller

import (
	"net/http"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	ws "golang.org/x/net/websocket"

	"github.com/neutrinocorp/boltzmann/controller/request"
	"github.com/neutrinocorp/boltzmann/controller/response"
	"github.com/neutrinocorp/boltzmann/scheduler"
	"github.com/neutrinocorp/boltzmann/state"
)

type TaskSchedulerHTTP struct {
	Service      scheduler.Service
	StateService state.Service
}

func (h TaskSchedulerHTTP) SetRoutes(g *echo.Group) {
	g.POST("/tasks/-/scheduler/schedule", h.schedule)
	g.GET("/tasks/:task_id", h.get)
	g.GET("/tasks/:task_id/ws", h.streamGet)
}

func (h TaskSchedulerHTTP) schedule(c echo.Context) error {
	req := request.ScheduleTasksRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	schedTasks, err := h.Service.Schedule(c.Request().Context(), req.Tasks)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Message{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, response.NewScheduledTasksResponse(schedTasks))
}

func (h TaskSchedulerHTTP) get(c echo.Context) error {
	taskID := c.Param("task_id")
	task, err := h.StateService.Get(c.Request().Context(), taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.NewContainerResponse(task))
}

func (h TaskSchedulerHTTP) streamGet(c echo.Context) error {
	handler := ws.Handler(func(conn *ws.Conn) {
		defer func() {
			if err := conn.Close(); err != nil {
				log.Err(err).Msg("failed to close ws connection")
			}
		}()

		connClose := atomic.Bool{}
		go func() {
			msg := ""
			if err := ws.Message.Receive(conn, &msg); err != nil && err.Error() == "EOF" {
				connClose.Store(true)
			}
		}()

		var taskHash string
		for {
			if connClose.Load() {
				break
			}

			task, newTaskHash, err := h.StateService.GetIfChanged(c.Request().Context(), taskHash,
				c.Param("task_id"))
			if err != nil {
				_ = ws.Message.Send(conn, err.Error())
				break
			} else if newTaskHash == taskHash {
				time.Sleep(time.Second * 5)
				continue
			}

			taskHash = newTaskHash
			taskJSON, err := jsoniter.Marshal(response.NewContainerResponse(task))
			if err != nil {
				_ = ws.Message.Send(conn, err.Error())
				break
			}
			err = ws.Message.Send(conn, string(taskJSON))
			if err != nil {
				log.Err(err).Msg("failed to write into ws stream")
			}

			time.Sleep(time.Second * 5)
		}
	})
	srv := ws.Server{
		Config: ws.Config{
			Location:  nil,
			Origin:    nil,
			Protocol:  nil,
			Version:   0,
			TlsConfig: nil,
			Header:    nil,
			Dialer:    nil,
		},
		Handshake: nil,
		Handler:   handler,
	}
	srv.ServeHTTP(c.Response(), c.Request())
	return nil
}
