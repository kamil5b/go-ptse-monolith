package domain

import "time"

// UserLoggedInEvent is published when a user successfully logs in
type UserLoggedInEvent struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	SessionID string    `json:"session_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	LoginAt   time.Time `json:"login_at"`
}

func (e UserLoggedInEvent) EventName() string { return "auth.user_logged_in" }
func (e UserLoggedInEvent) Payload() any      { return e }

// UserRegisteredEvent is published when a new user registers
type UserRegisteredEvent struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registered_at"`
}

func (e UserRegisteredEvent) EventName() string { return "auth.user_registered" }
func (e UserRegisteredEvent) Payload() any      { return e }

// UserLoggedOutEvent is published when a user logs out
type UserLoggedOutEvent struct {
	UserID     string    `json:"user_id"`
	SessionID  string    `json:"session_id,omitempty"`
	AllDevices bool      `json:"all_devices"`
	LogoutAt   time.Time `json:"logout_at"`
}

func (e UserLoggedOutEvent) EventName() string { return "auth.user_logged_out" }
func (e UserLoggedOutEvent) Payload() any      { return e }

// PasswordChangedEvent is published when a user changes their password
type PasswordChangedEvent struct {
	UserID    string    `json:"user_id"`
	ChangedAt time.Time `json:"changed_at"`
}

func (e PasswordChangedEvent) EventName() string { return "auth.password_changed" }
func (e PasswordChangedEvent) Payload() any      { return e }

// SessionRevokedEvent is published when a session is revoked
type SessionRevokedEvent struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	RevokedAt time.Time `json:"revoked_at"`
}

func (e SessionRevokedEvent) EventName() string { return "auth.session_revoked" }
func (e SessionRevokedEvent) Payload() any      { return e }
