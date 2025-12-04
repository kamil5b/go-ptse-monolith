package domain

import (
	"context"

	sharedctx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"
)

// Handler defines the interface for user HTTP handlers
type Handler interface {
	Create(c sharedctx.Context) error
	Get(c sharedctx.Context) error
	List(c sharedctx.Context) error
	Update(c sharedctx.Context) error
	Delete(c sharedctx.Context) error
}

// EmailSender defines the interface for sending user-related emails
type EmailSender interface {
	SendWelcomeEmail(ctx context.Context, userEmail, userName string) error
	SendPasswordResetEmail(ctx context.Context, userEmail, resetLink string) error
}

// Service defines the interface for user business logic
type Service interface {
	Create(ctx context.Context, req *CreateUserRequest, createdBy string) (*User, error)
	Get(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, req *UpdateUserRequest, updatedBy string) (*User, error)
	Delete(ctx context.Context, id, deletedBy string) error
}

// Repository defines the interface for user data access
type Repository interface {
	StartContext(ctx context.Context) context.Context
	DeferErrorContext(ctx context.Context, err error)

	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, u *User) error
	SoftDelete(ctx context.Context, id, deletedBy string) error
}
