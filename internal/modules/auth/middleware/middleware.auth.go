package middleware

import (
	"encoding/base64"
	"go-modular-monolith/internal/modules/auth/domain"
	sharedctx "go-modular-monolith/internal/shared/context"
	"net/http"
	"strings"
)

type AuthType string

const (
	AuthTypeJWT     AuthType = "jwt"
	AuthTypeSession AuthType = "session"
	AuthTypeBasic   AuthType = "basic"
	AuthTypeNone    AuthType = "none"
)

type MiddlewareConfig struct {
	AuthType       AuthType
	SkipPaths      []string
	SessionCookie  string
	BasicAuthRealm string
}

func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		AuthType:       AuthTypeJWT,
		SkipPaths:      []string{"/auth/login", "/auth/register", "/health"},
		SessionCookie:  "session_token",
		BasicAuthRealm: "Restricted",
	}
}

type AuthMiddleware struct {
	authService domain.Service
	config      MiddlewareConfig
}

func NewAuthMiddleware(authService domain.Service, config MiddlewareConfig) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		config:      config,
	}
}

func (m *AuthMiddleware) Authenticate() func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
	return func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
		return func(c sharedctx.Context) error {
			var authUser *domain.AuthUser
			var err error

			switch m.config.AuthType {
			case AuthTypeJWT:
				authUser, err = m.authenticateJWT(c)
			case AuthTypeSession:
				authUser, err = m.authenticateSession(c)
			case AuthTypeBasic:
				authUser, err = m.authenticateBasic(c)
			case AuthTypeNone:
				return next(c)
			default:
				authUser, err = m.authenticateJWT(c)
			}

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}

			if authUser != nil {
				c.Set("auth_user", authUser)
				c.Set("user_id", authUser.UserID)
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) RequireAuth() func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
	return func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
		return func(c sharedctx.Context) error {
			authUser := m.getAuthUser(c)
			if authUser == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			}
			return next(c)
		}
	}
}

func (m *AuthMiddleware) OptionalAuth() func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
	return func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
		return func(c sharedctx.Context) error {
			var authUser *domain.AuthUser

			switch m.config.AuthType {
			case AuthTypeJWT:
				authUser, _ = m.authenticateJWT(c)
			case AuthTypeSession:
				authUser, _ = m.authenticateSession(c)
			case AuthTypeBasic:
				authUser, _ = m.authenticateBasic(c)
			}

			if authUser != nil {
				c.Set("auth_user", authUser)
				c.Set("user_id", authUser.UserID)
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
	return func(next func(sharedctx.Context) error) func(sharedctx.Context) error {
		return func(c sharedctx.Context) error {
			authUser := m.getAuthUser(c)
			if authUser == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			}

			if !m.hasAnyRole(authUser.Roles, roles) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) authenticateJWT(c sharedctx.Context) (*domain.AuthUser, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, nil
	}

	token := parts[1]
	resp, err := m.authService.ValidateToken(c.GetContext(), token)
	if err != nil || !resp.Valid {
		return nil, err
	}

	return &domain.AuthUser{
		UserID:   resp.User.ID,
		Username: resp.User.Username,
		Email:    resp.User.Email,
		Roles:    resp.User.Roles,
		AuthType: domain.AuthTypeJWT,
	}, nil
}

func (m *AuthMiddleware) authenticateSession(c sharedctx.Context) (*domain.AuthUser, error) {
	sessionToken, err := c.GetCookie(m.config.SessionCookie)
	if err != nil || sessionToken == "" {
		return nil, nil
	}

	resp, err := m.authService.ValidateToken(c.GetContext(), sessionToken)
	if err != nil || !resp.Valid {
		return nil, err
	}

	return &domain.AuthUser{
		UserID:    resp.User.ID,
		Username:  resp.User.Username,
		Email:     resp.User.Email,
		Roles:     resp.User.Roles,
		SessionID: sessionToken,
		AuthType:  domain.AuthTypeSession,
	}, nil
}

func (m *AuthMiddleware) authenticateBasic(c sharedctx.Context) (*domain.AuthUser, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, nil
	}

	username := credentials[0]
	password := credentials[1]

	loginReq := &domain.LoginRequest{
		Username: username,
		Password: password,
	}

	resp, err := m.authService.Login(c.GetContext(), loginReq, c.GetUserAgent(), c.GetClientIP())
	if err != nil {
		return nil, err
	}

	return &domain.AuthUser{
		UserID:   resp.User.ID,
		Username: resp.User.Username,
		Email:    resp.User.Email,
		AuthType: domain.AuthTypeBasic,
	}, nil
}

func (m *AuthMiddleware) getAuthUser(c sharedctx.Context) *domain.AuthUser {
	val := c.Get("auth_user")
	if val == nil {
		return nil
	}
	if authUser, ok := val.(*domain.AuthUser); ok {
		return authUser
	}
	return nil
}

func (m *AuthMiddleware) hasAnyRole(userRoles, requiredRoles []string) bool {
	roleSet := make(map[string]bool)
	for _, r := range userRoles {
		roleSet[r] = true
	}
	for _, r := range requiredRoles {
		if roleSet[r] {
			return true
		}
	}
	return false
}
