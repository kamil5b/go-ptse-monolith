package http

import (
	"github.com/kamil5b/go-ptse-monolith/internal/app/core"
	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
	transportFiber "github.com/kamil5b/go-ptse-monolith/internal/transports/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func NewFiberServer(c *core.Container) *fiber.App {
	app := fiber.New()

	routes := NewRoutes(
		c.ProductHandler,
		c.UserHandler,
		c.AuthHandler,
		c.AuthMiddleware,
	)

	v1 := app.Group("/v1")
	for _, route := range *routes {
		switch h := route.Handler.(type) {
		case func(sharedctx.Context) error:
			finalHandler := applyMiddlewares(h, route.Middlewares)
			route.Handler = finalHandler
			transportFiber.AdapterToFiberRoutes(v1, &route, func(ctx *fiber.Ctx) sharedctx.Context {
				return transportFiber.NewFiberContext(ctx)
			})
		}
	}
	return app
}
