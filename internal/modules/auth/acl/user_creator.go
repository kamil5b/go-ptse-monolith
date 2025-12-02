package acl

import (
	"context"
	"time"

	"go-modular-monolith/internal/modules/auth/domain"
	userdomain "go-modular-monolith/internal/modules/user/domain"
)

// UserCreatorAdapter implements domain.UserCreator by adapting to the user module.
// This is the Anti-Corruption Layer (ACL) that translates between auth's needs
// and the user module's actual implementation.
type UserCreatorAdapter struct {
	userRepo userdomain.Repository
}

// NewUserCreatorAdapter creates a new ACL adapter for user creation.
func NewUserCreatorAdapter(userRepo userdomain.Repository) *UserCreatorAdapter {
	return &UserCreatorAdapter{
		userRepo: userRepo,
	}
}

// CreateUser implements domain.UserCreator interface.
// It translates auth's NewUser to user module's User domain model.
func (a *UserCreatorAdapter) CreateUser(ctx context.Context, newUser *domain.NewUser) error {
	// Translate from auth's view to user module's domain model
	user := &userdomain.User{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		CreatedAt: time.Now().UTC(),
		CreatedBy: newUser.CreatedBy,
	}

	return a.userRepo.Create(ctx, user)
}
