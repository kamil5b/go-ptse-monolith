package domain

import "time"

// UserCreatedEvent is published when a new user is created
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func (e UserCreatedEvent) EventName() string { return "user.created" }
func (e UserCreatedEvent) Payload() any      { return e }

// UserUpdatedEvent is published when a user is updated
type UserUpdatedEvent struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (e UserUpdatedEvent) EventName() string { return "user.updated" }
func (e UserUpdatedEvent) Payload() any      { return e }

// UserDeletedEvent is published when a user is soft-deleted
type UserDeletedEvent struct {
	UserID    string    `json:"user_id"`
	DeletedBy string    `json:"deleted_by"`
	DeletedAt time.Time `json:"deleted_at"`
}

func (e UserDeletedEvent) EventName() string { return "user.deleted" }
func (e UserDeletedEvent) Payload() any      { return e }
