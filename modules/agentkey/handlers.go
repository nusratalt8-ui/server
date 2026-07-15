package agentkey

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/apperr"
	"agentmanager/modules/auth"
)

func (s *Service) GetInfo(c echo.Context) error {
	uid := auth.UserIDFromContext(c)
	exists, createdAt := s.Info(uid)
	return c.JSON(http.StatusOK, echo.Map{"exists": exists, "created_at": createdAt})
}

func (s *Service) RotateKey(c echo.Context) error {
	uid := auth.UserIDFromContext(c)
	raw, err := s.Rotate(uid)
	if err != nil {
		return apperr.Internal(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{"key": raw})
}
