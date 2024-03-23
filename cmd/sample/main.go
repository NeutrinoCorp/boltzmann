package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func main() {
	e := echo.New()
	e.GET("/api/heartbeat", func(c echo.Context) error {
		log.Info().Str("header.authz", c.Request().Header.Get("Authorization")).Msg("got beat")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "beat",
		})
	})
	e.POST("/api/foo", func(c echo.Context) error {
		var data map[string]any
		if err := c.Bind(&data); err != nil {
			c.Error(err)
			return nil
		}
		log.Info().Any("data", data).Msg("got message")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "ok",
		})
	})
	if err := e.Start(":8081"); err != nil {
		panic(err)
	}
}
