package attachments

import "errors"

var (
	ErrTooLarge = errors.New("attachment too large")
	ErrNotFound = errors.New("attachment not found")
	ErrUpload   = errors.New("upload failed")
)
