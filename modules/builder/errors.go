package builder

import "errors"

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrNotFound         = errors.New("not found")
	ErrBuildNotFound    = errors.New("build not found")
	ErrNoIcon           = errors.New("no icon file")
	ErrIconType         = errors.New("must be .ico file")
	ErrIconTooLarge     = errors.New("icon too large, max 25MB")
	ErrPaidPlanRequired = errors.New("crypter requires paid plan")
	ErrBuildInProgress  = errors.New("build already in progress")
)
