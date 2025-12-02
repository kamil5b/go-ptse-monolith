package http

import (
	authdomain "go-modular-monolith/internal/modules/auth/domain"
	"go-modular-monolith/internal/modules/auth/middleware"
	productdomain "go-modular-monolith/internal/modules/product/domain"
	userdomain "go-modular-monolith/internal/modules/user/domain"
	"go-modular-monolith/internal/transports/http"
)

// MiddlewareFunc is a generic middleware function type
type MiddlewareFunc[T any] func(next func(T) error) func(T) error

func NewRoutes(
	productHandler productdomain.Handler,
	userHandler userdomain.Handler,
	authHandler authdomain.Handler,
	authMiddleware *middleware.AuthMiddleware,
) *[]http.Route {
	return &[]http.Route{
		// Auth routes (public - no middleware)
		{
			Method:  "POST",
			Path:    "/auth/login",
			Handler: authHandler.Login,
			Flags:   []string{"public"},
		},
		{
			Method:  "POST",
			Path:    "/auth/register",
			Handler: authHandler.Register,
			Flags:   []string{"public"},
		},
		{
			Method:  "POST",
			Path:    "/auth/refresh",
			Handler: authHandler.RefreshToken,
			Flags:   []string{"public"},
		},
		{
			Method:  "POST",
			Path:    "/auth/validate",
			Handler: authHandler.ValidateToken,
			Flags:   []string{"public"},
		},

		// Auth routes (protected - with auth middleware)
		{
			Method:      "POST",
			Path:        "/auth/logout",
			Handler:     authHandler.Logout,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "GET",
			Path:        "/auth/profile",
			Handler:     authHandler.GetProfile,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "PUT",
			Path:        "/auth/password",
			Handler:     authHandler.ChangePassword,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "GET",
			Path:        "/auth/sessions",
			Handler:     authHandler.GetSessions,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "DELETE",
			Path:        "/auth/sessions/:id",
			Handler:     authHandler.RevokeSession,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "DELETE",
			Path:        "/auth/sessions",
			Handler:     authHandler.RevokeAllSessions,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},

		// Product routes (can add middleware here if needed)
		{
			Method:      "GET",
			Path:        "/product",
			Handler:     productHandler.List,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "POST",
			Path:        "/product",
			Handler:     productHandler.Create,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},

		// User CRUD (can add middleware here if needed)
		{
			Method:      "GET",
			Path:        "/user",
			Handler:     userHandler.List,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "POST",
			Path:        "/user",
			Handler:     userHandler.Create,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "GET",
			Path:        "/user/:id",
			Handler:     userHandler.Get,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "PUT",
			Path:        "/user/:id",
			Handler:     userHandler.Update,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
		{
			Method:      "DELETE",
			Path:        "/user/:id",
			Handler:     userHandler.Delete,
			Middlewares: []any{authMiddleware.Authenticate(), authMiddleware.RequireAuth()},
			Flags:       []string{"protected"},
		},
	}
}
