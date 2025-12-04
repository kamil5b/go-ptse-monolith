package domain

import (
	"context"

	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

// Handler defines the interface for authentication HTTP handlers
type Handler interface {
	Login(c sharedctx.Context) error
	Register(c sharedctx.Context) error
	Logout(c sharedctx.Context) error
	RefreshToken(c sharedctx.Context) error
	ValidateToken(c sharedctx.Context) error
	ChangePassword(c sharedctx.Context) error
	GetProfile(c sharedctx.Context) error
	GetSessions(c sharedctx.Context) error
	RevokeSession(c sharedctx.Context) error
	RevokeAllSessions(c sharedctx.Context) error
}

// Service defines the interface for authentication business logic
type Service interface {
	// Authentication
	Login(ctx context.Context, req *LoginRequest, userAgent, ipAddress string) (*LoginResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Logout(ctx context.Context, userID string, req *LogoutRequest) error
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
	ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error)

	// Password management
	ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	ConfirmResetPassword(ctx context.Context, req *ConfirmResetPasswordRequest) error

	// Session management
	GetSessions(ctx context.Context, userID string) (*SessionListResponse, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID string) error

	// Token utilities
	GenerateAccessToken(claims *TokenClaims) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ParseToken(token string) (*TokenClaims, error)

	// Password utilities
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}

// Repository defines the interface for authentication data access
type Repository interface {
	StartContext(ctx context.Context) context.Context
	DeferErrorContext(ctx context.Context, err error)

	// Credential operations
	CreateCredential(ctx context.Context, cred *Credential) error
	GetCredentialByUsername(ctx context.Context, username string) (*Credential, error)
	GetCredentialByEmail(ctx context.Context, email string) (*Credential, error)
	GetCredentialByUserID(ctx context.Context, userID string) (*Credential, error)
	UpdateCredential(ctx context.Context, cred *Credential) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error

	// Session operations
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	GetSessionByID(ctx context.Context, id string) (*Session, error)
	GetSessionsByUserID(ctx context.Context, userID string) ([]Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	DeleteExpiredSessions(ctx context.Context) error
}

// Middleware defines the interface for authentication middleware
type Middleware interface {
	// Authenticate validates the request and sets auth context
	Authenticate() func(next func(sharedctx.Context) error) func(sharedctx.Context) error

	// RequireAuth ensures the request is authenticated
	RequireAuth() func(next func(sharedctx.Context) error) func(sharedctx.Context) error

	// OptionalAuth tries to authenticate but allows unauthenticated requests
	OptionalAuth() func(next func(sharedctx.Context) error) func(sharedctx.Context) error

	// RequireRoles ensures the authenticated user has specific roles
	RequireRoles(roles ...string) func(next func(sharedctx.Context) error) func(sharedctx.Context) error
}

// =============================================================================
// Anti-Corruption Layer (ACL) Interfaces
// These interfaces define contracts for external module dependencies.
// Auth module owns these interfaces, external modules provide implementations.
// =============================================================================

// UserCreator is an ACL interface for creating users during registration.
// This abstracts the dependency on the user module, allowing:
// - Clear contract definition owned by auth module
// - Easy testing with mocks
// - Swappable implementations for different scenarios
// - Future migration to event-driven or microservices architecture
type UserCreator interface {
	// CreateUser creates a new user account.
	// The auth module defines what it needs; the adapter translates to user module.
	CreateUser(ctx context.Context, user *NewUser) error
}

// NewUser represents the data auth module needs to create a user.
// This is auth's view of a user, not the user module's domain model.
type NewUser struct {
	ID        string
	Name      string
	Email     string
	CreatedBy string
}
