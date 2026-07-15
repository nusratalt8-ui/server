package agentkey

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/router"
)

const ctxUserID = "agent_user_id"

func (s *Service) Guard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := router.Bearer(c)
		userID, ok := s.Validate(token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": ErrInvalidKey.Error()})
		}
		c.Set(ctxUserID, userID)
		return next(c)
	}
}

// Tagger tags each agent websocket connection with the owning account id so
// the hub can scope broadcasts and resolve an agent's owner.
func Tagger(c echo.Context) string {
	if v, ok := c.Get(ctxUserID).(string); ok {
		return v
	}
	return ""
}
