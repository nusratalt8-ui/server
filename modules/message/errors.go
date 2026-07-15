package message

import "errors"

var (
	ErrEmpty   = errors.New("message has no content")
	ErrTooLong = errors.New("message too long")
)
