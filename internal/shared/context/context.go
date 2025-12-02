package context

import "context"

// Context defines the interface for HTTP context used by handlers
// This abstraction allows handlers to work with any HTTP framework (Echo, Gin, etc.)
type Context interface {
	// Binding methods
	BindJSON(obj any) error
	BindURI(obj any) error
	BindQuery(obj any) error
	BindHeader(obj any) error
	Bind(obj any) error

	// Response methods
	JSON(code int, v any) error

	// Request methods
	Param(name string) string
	GetUserID() string
	Get(key string) any
	Set(key string, value any)
	GetContext() context.Context

	// Header methods
	GetHeader(key string) string
	SetHeader(key, value string)

	// Cookie methods
	GetCookie(name string) (string, error)
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	RemoveCookie(name string)

	// Client info
	GetClientIP() string
	GetUserAgent() string
}
