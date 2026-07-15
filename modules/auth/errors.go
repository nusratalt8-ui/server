package auth

import (
	"errors"

	"agentmanager/modules/apperr"
)

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameTooLong  = errors.New("username is too long")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password is too short")
	ErrUsernameTaken    = errors.New("username is taken")
	ErrBadCredentials   = errors.New("invalid username or password")
	ErrRateLimited      = errors.New("too many requests")
	ErrInvalidSession   = errors.New("invalid session")
	ErrSessionExpired   = errors.New("session expired")
	ErrUnauthorized     = apperr.ErrUnauthorized
	ErrServerError      = apperr.ErrServerError
	ErrInvalidRequest   = apperr.ErrInvalidRequest
)
