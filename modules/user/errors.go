package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrNotAdmin     = errors.New("admin required")
)
