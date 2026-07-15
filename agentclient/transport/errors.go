package transport

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized: invalid or missing api token")
	ErrNotConnected = errors.New("not connected")
)
