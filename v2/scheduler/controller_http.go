package scheduler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann/v2/controller"
)

type ControllerHTTP struct {
	Service Service
}

var _ controller.VersionedHTTP = ControllerHTTP{}

func (h ControllerHTTP) SetRoutes(e *echo.Echo) {
	e.POST("/scheduler/schedule", h.schedule)
}

func (h ControllerHTTP) SetVersionedRoutes(_ *echo.Group) {
}

func (h ControllerHTTP) schedule(c echo.Context) error {
	cmd := ScheduleTasksCommand{}
	if err := c.Bind(&cmd); err != nil {
		return err
	}

	execPlan, err := h.Service.Schedule(c.Request().Context(), cmd)
	if err != nil {
		return err
	}

	responses := make([]controller.TaskResponse, 0, len(execPlan.Tasks))
	for _, task := range execPlan.Tasks {
		responses = append(responses, controller.TaskResponse{
			TaskID:          task.TaskID,
			ExecutionPlanID: task.ExecutionPlanID,
			Driver:          task.Driver,
			ResourceURL:     task.ResourceURL,
			AgentArguments:  task.AgentArguments,
			TypeMIME:        task.TypeMIME,
			PayloadSize:     len(task.EncodedPayload),
		})
	}

	return c.JSON(http.StatusOK, controller.ExecutionPlanResponse{
		PlanID:       execPlan.PlanID,
		Tasks:        responses,
		WithFairness: execPlan.WithFairness,
	})
}
