package http

import (
	"go-modular-monolith/internal/app/core"
	"go-modular-monolith/internal/domain/product"

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
	)
	for _, route := range *routes {
		switch route.Handler.(type) {
		case func(product.Context) error:
			v1 = transportEcho.AdapterToEchoRoutes(v1, &route, func(c echo.Context) product.Context {
				return transportEcho.NewEchoContext(c)
			}).Group("")
		}
	}
	return e
}
