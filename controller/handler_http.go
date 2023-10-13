package controller

import (
	"errors"
	"net/http"

	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"

	"github.com/neutrinocorp/boltzmann/controller/response"
)

var _ echo.HTTPErrorHandler = EchoErrHandler

func EchoErrHandler(err error, c echo.Context) {
	var multiErr *multierror.Error
	switch {
	case errors.As(err, &multiErr):
		msgs := make([]response.Message, 0, len(multiErr.Errors))
		for _, err := range multiErr.Errors {
			msgs = append(msgs, response.Message{Message: err.Error()})
		}
		_ = c.JSON(http.StatusInternalServerError, msgs)
	default:
		_ = c.JSON(http.StatusInternalServerError, response.Message{Message: err.Error()})
	}
}
