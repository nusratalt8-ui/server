package plan

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/admin"
	"agentmanager/modules/apperr"
)

func (s *Service) SetUserPlan(c echo.Context) error {
	uid := c.Get("auth_user_id").(string)
	if !admins.IsAdmin(uid) {
		return apperr.JSON(c, http.StatusForbidden, ErrNotAdmin)
	}
	targetID := c.Param("id")
	var req struct {
		Plan int `json:"plan"`
	}
	if err := c.Bind(&req); err != nil {
		return apperr.JSON(c, http.StatusBadRequest, ErrInvalidRequest)
	}
	if err := s.Set(targetID, req.Plan); err != nil {
		return apperr.Internal(c, err)
	}
	s.BroadcastPlanUpdate(targetID, req.Plan)
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}
