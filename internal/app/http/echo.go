package http

import (
	"go-modular-monolith/internal/app/core"
	sharedctx "go-modular-monolith/internal/shared/context"

	transportEcho "go-modular-monolith/internal/transports/http/echo"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEchoServer(c *core.Container) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	v1 := e.Group("/v1")
	routes := NewRoutes(
		c.ProductHandler,
		c.UserHandler,
		c.AuthHandler,
		c.AuthMiddleware,
	)

	for _, route := range *routes {
		switch h := route.Handler.(type) {
		case func(sharedctx.Context) error:
			// Apply middlewares if any
			finalHandler := applyMiddlewares(h, route.Middlewares)
			route.Handler = finalHandler
			v1 = transportEcho.AdapterToEchoRoutes(v1, &route, func(c echo.Context) sharedctx.Context {
				return transportEcho.NewEchoContext(c)
			}).Group("")
		}
	}
	return e
}
