package routes

type Route struct {
	Method      string
	Path        string
	Handler     any
	Middlewares []any
	Flags       []string
}
