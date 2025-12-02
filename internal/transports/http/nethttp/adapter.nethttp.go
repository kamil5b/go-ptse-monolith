package nethttp

import (
	"net/http"

	"go-modular-monolith/pkg/routes"

	"github.com/gorilla/mux"
)

func AdapterToNetHTTPRoutes[T any](
	r *mux.Router,
	route *routes.Route,
	domainContext func(http.ResponseWriter, *http.Request) T,
) *mux.Router {
	handler := func(w http.ResponseWriter, req *http.Request) {
		_ = route.Handler.(func(T) error)(domainContext(w, req))
	}

	// Convert method and register
	switch route.Method {
	case "GET":
		r.HandleFunc(convertPath(route.Path), handler).Methods("GET")
	case "POST":
		r.HandleFunc(convertPath(route.Path), handler).Methods("POST")
	case "PUT":
		r.HandleFunc(convertPath(route.Path), handler).Methods("PUT")
	case "PATCH":
		r.HandleFunc(convertPath(route.Path), handler).Methods("PATCH")
	case "DELETE":
		r.HandleFunc(convertPath(route.Path), handler).Methods("DELETE")
	default:
		r.HandleFunc(convertPath(route.Path), handler)
	}
	return r
}

// convertPath converts ":param" style to "{param}" for gorilla/mux
func convertPath(p string) string {
	// simple conversion: replace ":" with "{" and add "}" after param
	// e.g. /user/:id -> /user/{id}
	out := ""
	i := 0
	for i < len(p) {
		if p[i] == ':' {
			// copy '{'
			out += "{"
			i++
			// copy param name
			for i < len(p) && p[i] != '/' {
				out += string(p[i])
				i++
			}
			out += "}"
		} else {
			out += string(p[i])
			i++
		}
	}
	return out
}
