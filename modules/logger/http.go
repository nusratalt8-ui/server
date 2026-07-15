package logger

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func HTTPMiddleware(portID string) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogMethod:  true,
		LogURI:     true,
		LogLatency: true,
		LogError:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			line := fmt.Sprintf("%s%s%3d%s %s%-6s%s %s %s%s%s",
				bold, colorForStatus(v.Status), v.Status, reset,
				dim, v.Method, reset,
				v.URI,
				gray, v.Latency, reset)
			if v.Error != nil {
				line += " " + red + v.Error.Error() + reset
			}
			PortLog(portID, LevelInfo, "HTTP", line)
			return nil
		},
	})
}

func colorForStatus(s int) string {
	switch {
	case s >= 500:
		return red
	case s >= 400:
		return yellow
	case s >= 300:
		return cyan
	default:
		return green
	}
}
