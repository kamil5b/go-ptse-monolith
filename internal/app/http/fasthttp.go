package http

import (
	"go-modular-monolith/internal/app/core"
	sharedctx "go-modular-monolith/internal/shared/context"
	transportFast "go-modular-monolith/internal/transports/http/fasthttp"

	fasthttprouter "github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func NewFastHTTPServer(c *core.Container) fasthttp.RequestHandler {
	r := fasthttprouter.New()

	v1 := r.Group("/v1")
	routes := NewRoutes(
		c.ProductHandler,
		c.UserHandler,
		c.AuthHandler,
		c.AuthMiddleware,
	)

	for _, route := range *routes {
		switch h := route.Handler.(type) {
		case func(sharedctx.Context) error:
			finalHandler := applyMiddlewares(h, route.Middlewares)
			route.Handler = finalHandler
			transportFast.AdapterToFastHTTPRoutes(v1, &route, func(ctx *fasthttp.RequestCtx) sharedctx.Context {
				return transportFast.NewFastHTTPContext(ctx)
			})
		}
	}
	return r.Handler
}
