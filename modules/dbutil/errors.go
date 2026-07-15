package dbutil

import "errors"

var (
	ErrOpenFailed   = errors.New("failed to open database")
	ErrPingFailed   = errors.New("failed to reach database")
	ErrSchemaFailed = errors.New("failed to apply schema")
)
