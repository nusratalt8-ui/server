package auth

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"agentmanager/modules/admin"
	"agentmanager/modules/apperr"
	"agentmanager/modules/crypto"
	"agentmanager/modules/logger"
	"agentmanager/modules/turn"
)

var (
	loginMu       sync.Mutex
	loginLimiters = map[string]*rate.Limiter{}
)

func loginLimiter(ip string) *rate.Limiter {
	loginMu.Lock()
	defer loginMu.Unlock()
	if l, ok := loginLimiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(0.5, 5)
	loginLimiters[ip] = l
	return l
}

const (
	maxUsernameLen = 64
	minPasswordLen = 8
)

func (s *Service) Register(c echo.Context) error {
	ip := c.RealIP()
	if !loginLimiter(ip).Allow() {
		return apperr.JSON(c, http.StatusTooManyRequests, ErrRateLimited)
	}
	username, password, err := bindCredentials(c)
	if err != nil {
		return apperr.JSON(c, http.StatusBadRequest, err)
	}
	if username == "" {
		return apperr.JSON(c, http.StatusBadRequest, ErrUsernameRequired)
	}
	if len(username) > maxUsernameLen {
		return apperr.JSON(c, http.StatusBadRequest, ErrUsernameTooLong)
	}
	if password == "" {
		return apperr.JSON(c, http.StatusBadRequest, ErrPasswordRequired)
	}
	if len(password) < minPasswordLen {
		return apperr.JSON(c, http.StatusBadRequest, ErrPasswordTooShort)
	}
	if _, _, ok := s.user.GetCredentials(username); ok {
		return apperr.JSON(c, http.StatusConflict, ErrUsernameTaken)
	}
	hash, err := crypto.Hash(password)
	if err != nil {
		return apperr.Internal(c, err)
	}
	uid, err := s.user.Create(username, hash, time.Now().Unix())
	if err != nil {
		return apperr.Internal(c, err)
	}
	if s.onRegister != nil {
		if err := s.onRegister(uid); err != nil {
			return apperr.Internal(c, err)
		}
	}
	logger.Infof("[auth] Register OK ip=%s uid=%s", ip, uid)
	return s.issueSession(c, uid, username, time.Now())
}

func (s *Service) Login(c echo.Context) error {
	ip := c.RealIP()
	if !loginLimiter(ip).Allow() {
		logger.Infof("[auth] Login rate limited ip=%s", ip)
		return apperr.JSON(c, http.StatusTooManyRequests, ErrRateLimited)
	}
	username, password, err := bindCredentials(c)
	if err != nil {
		return apperr.JSON(c, http.StatusBadRequest, err)
	}
	uid, hash, ok := s.user.GetCredentials(username)
	if !ok || !crypto.Verify(hash, password) {
		logger.Infof("[auth] Login FAIL ip=%s", ip)
		return apperr.JSON(c, http.StatusUnauthorized, ErrBadCredentials)
	}
	logger.Infof("[auth] Login OK ip=%s uid=%s", ip, uid)
	return s.issueSession(c, uid, username, time.Now())
}

func (s *Service) Logout(c echo.Context) error {
	logger.Infof("[auth] Logout called")
	if token := s.tokenFromRequest(c); token != "" {
		logger.Infof("[auth] Logout deleting session tokenLen=%d", len(token))
		s.deleteSession(token)
	}
	s.clearCookie(c)
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func (s *Service) Me(c echo.Context) error {
	uid := UserIDFromContext(c)
	username := ""
	if u, ok := s.user.Get(uid); ok {
		username = u.Username
	}
	plan := 0
	if s.plan != nil {
		plan = s.plan.Get(uid)
	}
	logger.Infof("[auth] Me uid=%s username=%q plan=%d", uid, username, plan)
	return c.JSON(http.StatusOK, echo.Map{
		"user_id":  uid,
		"username": username,
		"is_admin": admins.IsAdmin(uid),
		"plan":     plan,
	})
}

func (s *Service) ICEServers(c echo.Context) error {
	return c.JSON(http.StatusOK, turn.ICEServers())
}

func (s *Service) GetUIPrefs(c echo.Context) error {
	uid := UserIDFromContext(c)
	raw := s.user.GetUIPrefs(uid)
	return c.JSONBlob(http.StatusOK, []byte(raw))
}

func (s *Service) SetUIPrefs(c echo.Context) error {
	uid := UserIDFromContext(c)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil || len(body) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid body"})
	}
	if err := s.user.SetUIPrefs(uid, string(body)); err != nil {
		return apperr.Internal(c, err)
	}
	return c.JSONBlob(http.StatusOK, body)
}

func bindCredentials(c echo.Context) (string, string, error) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return "", "", ErrInvalidRequest
	}
	return strings.TrimSpace(req.Username), req.Password, nil
}

func (s *Service) issueSession(c echo.Context, uid, username string, now time.Time) error {
	token := crypto.RandomHex(32)
	if token == "" {
		return apperr.JSON(c, http.StatusInternalServerError, ErrServerError)
	}
	expires := now.Add(s.cfg.SessionTTL)
	if err := s.createSession(token, uid, now.Unix(), expires.Unix()); err != nil {
		logger.Infof("[auth] issueSession createSession ERR: %v", err)
		return apperr.Internal(c, err)
	}
	logger.Infof("[auth] issueSession OK uid=%s tokenLen=%d", uid, len(token))
	s.setCookie(c, token, expires)
	return c.JSON(http.StatusOK, echo.Map{
		"user_id":  uid,
		"username": username,
		"token":    token,
	})
}
