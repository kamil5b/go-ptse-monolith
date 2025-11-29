package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func JWTMiddleware(s *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Request().Header.Get("Authorization")
			if h == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing auth")
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid auth header")
			}
			claims, err := s.ParseToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			if sub, ok := claims["sub"].(string); ok {
				c.Set("user_id", sub)
			}
			return next(c)
		}
	}
}
