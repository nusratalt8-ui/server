package loader

import "errors"

var (
	ErrLoadFailed   = errors.New("plugin load failed")
	ErrProcMissing  = errors.New("plugin export not found")
	ErrInjectFailed = errors.New("injection failed")
)
