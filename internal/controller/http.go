package controller

import "github.com/labstack/echo/v4"

// HTTP is the interface for controllers based on the HTTP protocol.
type HTTP interface {
	// SetRoutes allocates routes into the given echo.Echo instance.
	SetRoutes(e *echo.Echo)
}

// VersionedHTTP is an extension of HTTP interface.
// It allows controllers to register routes in paths with a versioning prefix (e.g., /v1/foo, /v10/bar).
type VersionedHTTP interface {
	HTTP
	// SetVersionedRoutes registers routes in paths with a versioning prefix (e.g., /v1/foo, /v10/bar)
	SetVersionedRoutes(g *echo.Group)
}
