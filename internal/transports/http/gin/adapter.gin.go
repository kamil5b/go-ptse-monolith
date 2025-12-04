package gin

import (
	transportHTTP "github.com/kamil5b/go-ptse-monolith/internal/transports/http"

	"github.com/gin-gonic/gin"
)

func AdapterToGinRoutes[T any](
	r *gin.RouterGroup,
	route *transportHTTP.Route,
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
