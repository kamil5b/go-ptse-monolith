package http

import (
	"net/http"

	"go-modular-monolith/internal/app/core"
	sharedctx "go-modular-monolith/internal/shared/context"
	transportNet "go-modular-monolith/internal/transports/http/nethttp"

	"github.com/gorilla/mux"
)

func NewNetHTTPServer(c *core.Container) http.Handler {
	r := mux.NewRouter()

	v1 := r.PathPrefix("/v1").Subrouter()
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
			transportNet.AdapterToNetHTTPRoutes(v1, &route, func(w http.ResponseWriter, r *http.Request) sharedctx.Context {
				return transportNet.NewNetHTTPContext(w, r)
			})
		}
	}
	return r
}
