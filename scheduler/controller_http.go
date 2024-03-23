package scheduler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann/execplan"
	"github.com/neutrinocorp/boltzmann/internal/controller"
)

// ControllerHTTP is the Scheduler controller for HTTP transport protocol.
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

	return c.JSON(http.StatusOK, controller.NewViewData[execplan.ExecutionPlan](execPlan.View()))
}
