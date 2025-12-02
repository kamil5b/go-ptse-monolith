package domain

import "time"

// Session represents a user session stored in the database
type Session struct {
	ID        string     `db:"id" json:"id" bson:"id"`
	UserID    string     `db:"user_id" json:"user_id" bson:"user_id"`
	Token     string     `db:"token" json:"token" bson:"token"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at" bson:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	RevokedAt *time.Time `db:"revoked_at" json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
	UserAgent string     `db:"user_agent" json:"user_agent" bson:"user_agent"`
	IPAddress string     `db:"ip_address" json:"ip_address" bson:"ip_address"`
}

// Credential represents user credentials for authentication
type Credential struct {
	ID           string     `db:"id" json:"id" bson:"id"`
	UserID       string     `db:"user_id" json:"user_id" bson:"user_id"`
	Username     string     `db:"username" json:"username" bson:"username"`
	Email        string     `db:"email" json:"email" bson:"email"`
	PasswordHash string     `db:"password_hash" json:"-" bson:"password_hash"`
	IsActive     bool       `db:"is_active" json:"is_active" bson:"is_active"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at" bson:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles,omitempty"`
}

// AuthUser represents the authenticated user info extracted from auth context
type AuthUser struct {
	UserID    string
	Username  string
	Email     string
	Roles     []string
	SessionID string // For session-based auth
	AuthType  AuthType
}

// AuthType represents the type of authentication used
type AuthType string

const (
	AuthTypeJWT     AuthType = "jwt"
	AuthTypeSession AuthType = "session"
	AuthTypeBasic   AuthType = "basic"
	AuthTypeNone    AuthType = "none"
)
