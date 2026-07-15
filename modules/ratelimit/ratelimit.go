package ratelimit

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"agentmanager/modules/apperr"
)

func PerIP(cfg Config) echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      rate.Limit(cfg.PerSecond),
		Burst:     cfg.Burst,
		ExpiresIn: 3 * time.Minute,
	})
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: store,
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return apperr.JSON(c, http.StatusTooManyRequests, ErrRateLimited)
		},
	})
}
