package http

import (
	"go-modular-monolith/internal/app/core"
	sharedctx "go-modular-monolith/internal/shared/context"
	transportFiber "go-modular-monolith/internal/transports/http/fiber"

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
