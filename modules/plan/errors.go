package plan

import "errors"

var (
	ErrNotAdmin       = errors.New("admin required")
	ErrInvalidRequest = errors.New("invalid request")
)
