package apperr

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrServerError    = errors.New("server error")
	ErrNotFound       = errors.New("not found")
	ErrNotAllowed     = errors.New("not allowed")
	ErrUnauthorized   = errors.New("unauthorized")
)
