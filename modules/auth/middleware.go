package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"agentmanager/modules/apperr"
	"agentmanager/modules/logger"
	"agentmanager/modules/router"
)

func (s *Service) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := s.tokenFromRequest(c)
		if token == "" {
			return apperr.JSON(c, http.StatusUnauthorized, ErrUnauthorized)
		}
		row, ok := s.lookupSession(token)
		if !ok {
			return apperr.JSON(c, http.StatusUnauthorized, ErrInvalidSession)
		}
		if row.expiresAt < time.Now().Unix() {
			s.deleteSession(token)
			return apperr.JSON(c, http.StatusUnauthorized, ErrSessionExpired)
		}
		c.Set(ctxUserID, row.userID)
		return next(c)
	}
}

func (s *Service) tokenFromRequest(c echo.Context) string {
	if t := router.Bearer(c); t != "" {
		return t
	}
	if cookie, err := c.Cookie(s.cfg.CookieName); err == nil {
		return cookie.Value
	}
	return ""
}

func (s *Service) setCookie(c echo.Context, token string, expires time.Time) {
	logger.Infof("[auth] setCookie name=%q secure=%v exp=%s", s.cfg.CookieName, s.cfg.Secure, expires)
	c.SetCookie(&http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   s.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) clearCookie(c echo.Context) {
	logger.Infof("[auth] clearCookie name=%q", s.cfg.CookieName)
	c.SetCookie(&http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func UserIDFromContext(c echo.Context) string {
	if v, ok := c.Get(ctxUserID).(string); ok {
		return v
	}
	return ""
}
