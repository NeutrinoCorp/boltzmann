package task

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann"
	"github.com/neutrinocorp/boltzmann/internal/controller"
	"github.com/neutrinocorp/boltzmann/state"
)

type ControllerHTTP struct {
	Service state.Service[boltzmann.Task]
}

var _ controller.HTTP = ControllerHTTP{}

func (h ControllerHTTP) SetRoutes(e *echo.Echo) {
	e.GET("/states/plans/:planId/tasks/:taskId", h.getByID)
}

func (h ControllerHTTP) getByID(c echo.Context) error {
	planId := c.Param("planId")
	taskId, err := strconv.Atoi(c.Param("taskId"))
	if err != nil {
		return err
	}

	task := boltzmann.Task{
		ExecutionPlanID: planId,
		TaskID:          taskId,
	}

	st, err := h.Service.Get(c.Request().Context(), task.GetID())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, controller.NewViewData[state.State](st.View()))
}
