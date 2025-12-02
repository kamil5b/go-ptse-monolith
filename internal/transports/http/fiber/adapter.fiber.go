package fiber

import (
	"go-modular-monolith/pkg/routes"

	fiberpkg "github.com/gofiber/fiber/v2"
)

func AdapterToFiberRoutes[T any](
	r fiberpkg.Router,
	route *routes.Route,
	domainContext func(*fiberpkg.Ctx) T,
) {
	handler := func(c *fiberpkg.Ctx) error {
		return route.Handler.(func(T) error)(domainContext(c))
	}

	switch route.Method {
	case "GET":
		r.Get(route.Path, handler)
	case "POST":
		r.Post(route.Path, handler)
	case "PUT":
		r.Put(route.Path, handler)
	case "PATCH":
		r.Patch(route.Path, handler)
	case "DELETE":
		r.Delete(route.Path, handler)
	default:
		r.All(route.Path, handler)
	}
}
