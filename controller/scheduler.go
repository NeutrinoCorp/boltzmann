package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann/controller/request"
	"github.com/neutrinocorp/boltzmann/controller/response"
	"github.com/neutrinocorp/boltzmann/scheduler"
)

type TaskSchedulerHTTP struct {
	Service scheduler.Service
}

func (h TaskSchedulerHTTP) SetRoutes(g *echo.Group) {
	g.POST("/tasks/-/scheduler/schedule", h.schedule)
	g.GET("/tasks/:task_id", h.get)
}

func (h TaskSchedulerHTTP) schedule(c echo.Context) error {
	req := request.ScheduleTasksRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	schedTasks := h.Service.Schedule(c.Request().Context(), req.Tasks)
	return c.JSON(http.StatusOK, response.NewScheduledTasksResponse(schedTasks))
}

func (h TaskSchedulerHTTP) get(c echo.Context) error {
	taskID := c.Param("task_id")
	task, err := h.Service.GetTaskState(c.Request().Context(), taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.NewContainerResponse(task))
}
