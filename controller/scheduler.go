package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/command"
	"github.com/neutrinocorp/boltzmann/scheduler"
)

type TaskSchedulerHTTP struct {
	Service scheduler.Service
}

func (h TaskSchedulerHTTP) SetRoutes(g *echo.Group) {
	g.POST("/tasks/-/scheduler/schedule", h.schedule)
	g.GET("/tasks/:task_id", h.get)
}

type ScheduleTasksRequest struct {
	Tasks []command.ScheduleTaskCommand `json:"tasks"`
}

type ScheduledTaskResponse struct {
	TaskID        string    `json:"task_id"`
	CorrelationID string    `json:"correlation_id"`
	Driver        string    `json:"driver"`
	ResourceURI   string    `json:"resource_uri"`
	ErrorMessage  string    `json:"error_message"`
	ScheduleTime  time.Time `json:"schedule_time"`
}

type ScheduleTasksResponse struct {
	Tasks []ScheduledTaskResponse `json:"tasks"`
}

func (h TaskSchedulerHTTP) schedule(c echo.Context) error {
	req := ScheduleTasksRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	schedTasks := h.Service.Schedule(c.Request().Context(), req.Tasks)
	res := ScheduleTasksResponse{
		Tasks: make([]ScheduledTaskResponse, 0, len(schedTasks)),
	}
	for _, result := range schedTasks {
		res.Tasks = append(res.Tasks, ScheduledTaskResponse{
			TaskID:        result.TaskID,
			CorrelationID: result.CorrelationID,
			Driver:        result.Driver,
			ResourceURI:   result.ResourceURI,
			ErrorMessage:  result.ErrorMessage,
			ScheduleTime:  result.ScheduleTime,
		})
	}

	return c.JSON(http.StatusOK, res)
}

type TaskResponse struct {
	Task boltzmann.Task `json:"task"`
}

func (h TaskSchedulerHTTP) get(c echo.Context) error {
	taskID := c.Param("task_id")
	task, err := h.Service.GetTaskState(c.Request().Context(), taskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, TaskResponse{Task: task})
}
