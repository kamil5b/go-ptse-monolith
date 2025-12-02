package domain

import "time"

// LoginResponse represents the login response payload
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *UserInfo `json:"user,omitempty"`
}

// UserInfo represents basic user info in auth responses
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Name     string   `json:"name,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

// RegisterResponse represents the registration response payload
type RegisterResponse struct {
	User    *UserInfo `json:"user"`
	Message string    `json:"message"`
}

// RefreshTokenResponse represents the refresh token response payload
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ValidateTokenResponse represents the validate token response payload
type ValidateTokenResponse struct {
	Valid  bool      `json:"valid"`
	User   *UserInfo `json:"user,omitempty"`
	Reason string    `json:"reason,omitempty"`
}

// SessionInfo represents session information
type SessionInfo struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Current   bool      `json:"current"`
}

// SessionListResponse represents the list of active sessions
type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}
