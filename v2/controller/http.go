package controller

import "github.com/labstack/echo/v4"

type HTTP interface {
	SetRoutes(e *echo.Echo)
}

type VersionedHTTP interface {
	HTTP
	SetVersionedRoutes(g *echo.Group)
}
