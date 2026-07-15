package agents

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/admin"
	"agentmanager/modules/apperr"
	"agentmanager/modules/auth"
	"agentmanager/modules/pagination"
)

func (s *Service) ListAgents(c echo.Context) error {
	uid := auth.UserIDFromContext(c)
	offset, limit := pagination.ParseOffsetLimit(c, agentsPageSize, pagination.MaxLimit)
	items, err := s.listPaged(uid, limit, offset)
	if err != nil {
		return apperr.Internal(c, err)
	}
	total, online, err := s.countsForUser(uid)
	if err != nil {
		return apperr.Internal(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{
		"items":        items,
		"total":        total,
		"online_total": online,
		"offset":       offset,
		"limit":        limit,
		"has_more":     offset+len(items) < total,
	})
}

func (s *Service) ListAllAgents(c echo.Context) error {
	uid := auth.UserIDFromContext(c)
	if !admins.IsAdmin(uid) {
		return apperr.JSON(c, http.StatusForbidden, ErrForbidden)
	}
	offset, limit := pagination.ParseOffsetLimit(c, agentsPageSize, pagination.MaxLimit)
	items, err := s.listAllPaged(limit, offset)
	if err != nil {
		return apperr.Internal(c, err)
	}
	perUser, totalAll, onlineAll, err := s.countsPerUser()
	if err != nil {
		return apperr.Internal(c, err)
	}
	return c.JSON(http.StatusOK, echo.Map{
		"items":        items,
		"total":        totalAll,
		"online_total": onlineAll,
		"offset":       offset,
		"limit":        limit,
		"has_more":     offset+len(items) < totalAll,
		"per_user":     perUser,
	})
}

func (s *Service) ListUserAgents(c echo.Context) error {
	uid := auth.UserIDFromContext(c)
	if !admins.IsAdmin(uid) {
		return apperr.JSON(c, http.StatusForbidden, ErrForbidden)
	}
	targetUID := c.Param("uid")
	if targetUID == "" {
		return apperr.NotFound(c)
	}
	offset, limit := pagination.ParseOffsetLimit(c, agentsPageSize, pagination.MaxLimit)
	items, err := s.listByUserPaged(targetUID, limit, offset)
	if err != nil {
		return apperr.Internal(c, err)
	}
	counts, ok := func() (UserCounts, bool) {
		perUser, _, _, err := s.countsPerUser()
		if err != nil {
			return UserCounts{}, false
		}
		c, ok := perUser[targetUID]
		return c, ok
	}()
	if !ok {
		counts = UserCounts{Total: len(items)}
	}
	return c.JSON(http.StatusOK, echo.Map{
		"items":        items,
		"total":        counts.Total,
		"online_total": counts.Online,
		"offset":       offset,
		"limit":        limit,
		"has_more":     offset+len(items) < counts.Total,
	})
}

func (s *Service) GetAgent(c echo.Context) error {
	agentID := c.Param("id")
	uid := auth.UserIDFromContext(c)
	a, ok := s.getByID(agentID)
	if !ok {
		return apperr.NotFound(c)
	}
	if a.UserID != uid && !admins.IsAdmin(uid) {
		return apperr.NotFound(c)
	}
	return c.JSON(http.StatusOK, a)
}

func (s *Service) DeleteAgent(c echo.Context) error {
	agentID := c.Param("id")
	uid := auth.UserIDFromContext(c)
	a, ok := s.getByID(agentID)
	if !ok {
		return apperr.NotFound(c)
	}
	if a.UserID != uid && !admins.IsAdmin(uid) {
		return apperr.NotFound(c)
	}
	if err := s.deleteAgent(agentID); err != nil {
		return apperr.Internal(c, err)
	}
	s.broadcast(a.UserID)
	return c.NoContent(http.StatusNoContent)
}
