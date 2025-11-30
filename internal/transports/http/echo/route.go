package echo

import (
	"go-modular-monolith/pkg/routes"

	"github.com/labstack/echo/v4"
)

func AppRoutesToEchoRoutes(
	e *echo.Group,
	routes *routes.Route,
) *echo.Group {
	switch routes.Method {
	case "GET":
		e.GET(routes.Path, func(ctx echo.Context) error {
			return routes.Handler.(func(EchoContext) error)(NewEchoContext(ctx))
		})
	case "POST":
		e.POST(routes.Path, func(ctx echo.Context) error {
			return routes.Handler.(func(EchoContext) error)(NewEchoContext(ctx))
		})
	case "PUT":
		e.PUT(routes.Path, func(ctx echo.Context) error {
			return routes.Handler.(func(EchoContext) error)(NewEchoContext(ctx))
		})
	case "PATCH":
		e.PATCH(routes.Path, func(ctx echo.Context) error {
			return routes.Handler.(func(EchoContext) error)(NewEchoContext(ctx))
		})
	case "DELETE":
		e.DELETE(routes.Path, func(ctx echo.Context) error {
			return routes.Handler.(func(EchoContext) error)(NewEchoContext(ctx))
		})
	}
	return e
}
