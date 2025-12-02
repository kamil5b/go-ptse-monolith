package http

// Route defines a framework-agnostic route registration
type Route struct {
	Method      string
	Path        string
	Handler     any
	Middlewares []any
	Flags       []string // Feature flags that control this route
}

// RouteGroup represents a group of routes with a common prefix
type RouteGroup struct {
	Prefix      string
	Routes      []Route
	Middlewares []any
}
