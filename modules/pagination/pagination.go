package pagination

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const DefaultLimit = 50
const MaxLimit = 200

func ParseOffsetLimit(c echo.Context, defaultLimit, maxLimit int) (offset, limit int) {
	limit = defaultLimit
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			limit = n
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}
