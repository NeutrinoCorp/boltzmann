package execplan

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann/v2"
	"github.com/neutrinocorp/boltzmann/v2/controller"
	"github.com/neutrinocorp/boltzmann/v2/state"
)

type ControllerHTTP struct {
	Service state.Service[boltzmann.ExecutionPlanReference]
}

var _ controller.HTTP = ControllerHTTP{}

func (h ControllerHTTP) SetRoutes(e *echo.Echo) {
	e.GET("/plans/:planId", h.getByID)
}

func (h ControllerHTTP) getByID(c echo.Context) error {
	planId := c.Param("planId")

	plan := boltzmann.ExecutionPlan{
		PlanID: planId,
	}

	st, err := h.Service.Get(c.Request().Context(), plan.GetID())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, st)
}
