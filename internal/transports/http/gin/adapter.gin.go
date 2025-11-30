package gin

import (
	"go-modular-monolith/pkg/routes"

	"github.com/gin-gonic/gin"
)

func AdapterToGinRoutes[T any](
	r *gin.RouterGroup,
	route *routes.Route,
	domainContext func(*gin.Context) T,
) *gin.RouterGroup {
	handler := func(ctx *gin.Context) {
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
	}
	return r
}
