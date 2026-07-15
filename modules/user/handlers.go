package user

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/admin"
	"agentmanager/modules/apperr"
)

func (s *Service) GetProfile(c echo.Context) error {
	u, ok := s.Get(c.Param("id"))
	if !ok {
		return apperr.JSON(c, http.StatusNotFound, ErrUserNotFound)
	}
	return c.JSON(http.StatusOK, u)
}

func (s *Service) ListUsers(c echo.Context) error {
	uid := c.Get("auth_user_id").(string)
	if !admins.IsAdmin(uid) {
		return apperr.JSON(c, http.StatusForbidden, ErrNotAdmin)
	}
	users, err := s.ListAll()
	if err != nil {
		return apperr.Internal(c, err)
	}
	return c.JSON(http.StatusOK, users)
}
