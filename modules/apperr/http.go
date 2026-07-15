package apperr

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/logger"
)

func JSON(c echo.Context, status int, err error) error {
	return c.JSON(status, echo.Map{"error": err.Error()})
}

func Internal(c echo.Context, err error) error {
	logger.Errorf("%s %s -> %v", c.Request().Method, c.Request().URL.Path, err)
	return c.JSON(http.StatusInternalServerError, echo.Map{"error": ErrServerError.Error()})
}

func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, echo.Map{"error": ErrNotFound.Error()})
}
