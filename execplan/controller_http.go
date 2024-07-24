package execplan

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann/internal/controller"
	"github.com/neutrinocorp/boltzmann/state"
)

type ControllerHTTP struct {
	Service      Service
	StateService state.Service[ExecutionPlanReference]
}

var _ controller.HTTP = ControllerHTTP{}

func (h ControllerHTTP) SetRoutes(e *echo.Echo) {
	e.GET("/states/plans/:planId", h.getStateByID)
	e.GET("/plans/:planId", h.getByID)
}

func (h ControllerHTTP) getStateByID(c echo.Context) error {
	planId := c.Param("planId")
	plan := ExecutionPlan{
		PlanID: planId,
	}

	st, err := h.StateService.Get(c.Request().Context(), plan.GetID())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, controller.NewViewData[state.State](st.View()))
}

func (h ControllerHTTP) getByID(c echo.Context) error {
	planId := c.Param("planId")
	plan := ExecutionPlan{
		PlanID: planId,
	}

	plan, err := h.Service.FindByID(c.Request().Context(), plan.GetID())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, controller.NewViewData[ExecutionPlan](plan.View()))
}
