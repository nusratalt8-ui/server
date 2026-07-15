package message

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/admin"
	"agentmanager/modules/apperr"
	"agentmanager/modules/auth"
)

func (s *Service) History(c echo.Context) error {
	agentID := c.Param("agent")
	if agentID == "" {
		return apperr.NotFound(c)
	}
	uid := auth.UserIDFromContext(c)
	isAdmin := admins.IsAdmin(uid)
	if !isAdmin {
		if owner, ok := s.reg.OwnerOf(agentID); !ok || owner != uid {
			return apperr.NotFound(c)
		}
	}
	before := c.QueryParam("before")
	limit := s.cfg.DefaultLimit
	list, err := s.fetchMessages(agentID, before, limit)
	if err != nil {
		return apperr.Internal(c, err)
	}
	if !isAdmin {
		filtered := list[:0]
		for _, m := range list {
			if m.Sender != "admin" && m.Sender != "agent_admin" {
				filtered = append(filtered, m)
			}
		}
		list = filtered
	}
	return c.JSON(http.StatusOK, echo.Map{"messages": list, "has_more": len(list) == limit})
}
