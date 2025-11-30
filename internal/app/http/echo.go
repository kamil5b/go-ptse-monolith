package http

import (
	"go-modular-monolith/internal/app/core"

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
		v1 = transportEcho.AppRoutesToEchoRoutes(v1, &route).Group("")
	}
	return e
}
