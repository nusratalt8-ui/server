package router

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type RouteConfig struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
	Auth    bool
	CSRF    bool
}

type Module interface {
	Prefix() string
	Routes() []RouteConfig
}

type Router struct {
	echo      *echo.Echo
	apiPrefix string
	authMw    echo.MiddlewareFunc
	csrfMw    echo.MiddlewareFunc
}

func New(e *echo.Echo, apiPrefix string, authMw, csrfMw echo.MiddlewareFunc) *Router {
	return &Router{echo: e, apiPrefix: apiPrefix, authMw: authMw, csrfMw: csrfMw}
}

func (r *Router) APIPrefix() string { return r.apiPrefix }

func (r *Router) RegisterModule(m Module) {
	prefix := r.apiPrefix + m.Prefix()
	for _, rt := range m.Routes() {
		rt.Path = prefix + rt.Path
		r.Register(rt)
	}
}

func (r *Router) Register(cfg RouteConfig) {
	var mw []echo.MiddlewareFunc
	if cfg.Auth && r.authMw != nil {
		mw = append(mw, r.authMw)
	}
	if cfg.CSRF && r.csrfMw != nil {
		mw = append(mw, r.csrfMw)
	}
	switch cfg.Method {
	case http.MethodGet:
		r.echo.GET(cfg.Path, cfg.Handler, mw...)
	case http.MethodPost:
		r.echo.POST(cfg.Path, cfg.Handler, mw...)
	case http.MethodPut:
		r.echo.PUT(cfg.Path, cfg.Handler, mw...)
	case http.MethodPatch:
		r.echo.PATCH(cfg.Path, cfg.Handler, mw...)
	case http.MethodDelete:
		r.echo.DELETE(cfg.Path, cfg.Handler, mw...)
	}
}

func (r *Router) Static(path, dir string) {
	r.echo.Static(path, dir)
}

func Bearer(c echo.Context) string {
	h := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(h[len("Bearer "):])
	}
	return ""
}
