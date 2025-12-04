package echo

import (
	transportHTTP "github.com/kamil5b/go-ptse-monolith/internal/transports/http"

	"github.com/labstack/echo/v4"
)

func AdapterToEchoRoutes[T any](
	e *echo.Group,
	route *transportHTTP.Route,
	domainContext func(echo.Context) T,
) *echo.Group {
	handler := func(ctx echo.Context) error {
		return route.Handler.(func(T) error)(domainContext(ctx))
	}

	switch route.Method {
	case echo.GET:
		e.GET(route.Path, handler)
	case echo.POST:
		e.POST(route.Path, handler)
	case echo.PUT:
		e.PUT(route.Path, handler)
	case echo.PATCH:
		e.PATCH(route.Path, handler)
	case echo.DELETE:
		e.DELETE(route.Path, handler)
	}
	return e
}
