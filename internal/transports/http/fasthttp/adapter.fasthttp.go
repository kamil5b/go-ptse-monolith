package fasthttp

import (
	"go-modular-monolith/pkg/routes"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func AdapterToFastHTTPRoutes[T any](
	r *router.Group,
	route *routes.Route,
	domainContext func(*fasthttp.RequestCtx) T,
) *router.Group {
	handler := func(ctx *fasthttp.RequestCtx) {
		_ = route.Handler.(func(T) error)(domainContext(ctx))
	}

	switch route.Method {
	case "GET":
		r.GET(route.Path, handler)
	case "POST":
		r.POST(route.Path, handler)
	case "PUT":
		r.PUT(route.Path, handler)
	case "PATCH":
		r.PATCH(route.Path, handler)
	case "DELETE":
		r.DELETE(route.Path, handler)
	default:
		r.Handle(route.Method, route.Path, handler)
	}
	return r
}
