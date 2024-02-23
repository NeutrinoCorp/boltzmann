package scheduler

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/neutrinocorp/boltzmann/v2/controller"
)

type ControllerHTTP struct {
	Service Service
}

var _ controller.HTTP = ControllerHTTP{}

func (h ControllerHTTP) SetRoutes(e *echo.Echo) {
	e.POST("/scheduler/schedule", h.schedule)
}

func (h ControllerHTTP) schedule(c echo.Context) error {
	cmd := ScheduleTasksCommand{}
	if err := c.Bind(&cmd); err != nil {
		return err
	}

	if err := h.Service.Schedule(c.Request().Context(), cmd); err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}
