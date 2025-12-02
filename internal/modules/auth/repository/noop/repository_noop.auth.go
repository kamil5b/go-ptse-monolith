package noop

import (
	"context"
	"errors"
	"go-modular-monolith/internal/modules/auth/domain"
)

var ErrNotImplemented = errors.New("auth repository not implemented")

type NoopRepository struct{}

func NewNoopRepository() *NoopRepository {
	return &NoopRepository{}
}

func (r *NoopRepository) StartContext(ctx context.Context) context.Context {
	return ctx
}

func (r *NoopRepository) DeferErrorContext(ctx context.Context, err error) {}

// Credential operations

func (r *NoopRepository) CreateCredential(ctx context.Context, cred *domain.Credential) error {
	return ErrNotImplemented
}

func (r *NoopRepository) GetCredentialByUsername(ctx context.Context, username string) (*domain.Credential, error) {
	return nil, ErrNotImplemented
}

func (r *NoopRepository) GetCredentialByEmail(ctx context.Context, email string) (*domain.Credential, error) {
	return nil, ErrNotImplemented
}

func (r *NoopRepository) GetCredentialByUserID(ctx context.Context, userID string) (*domain.Credential, error) {
	return nil, ErrNotImplemented
}

func (r *NoopRepository) UpdateCredential(ctx context.Context, cred *domain.Credential) error {
	return ErrNotImplemented
}

func (r *NoopRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return ErrNotImplemented
}

func (r *NoopRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	return ErrNotImplemented
}

// Session operations

func (r *NoopRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	return ErrNotImplemented
}

func (r *NoopRepository) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	return nil, ErrNotImplemented
}

func (r *NoopRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	return nil, ErrNotImplemented
}

func (r *NoopRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	return nil, ErrNotImplemented
}

func (r *NoopRepository) RevokeSession(ctx context.Context, sessionID string) error {
	return ErrNotImplemented
}

func (r *NoopRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return ErrNotImplemented
}

func (r *NoopRepository) DeleteExpiredSessions(ctx context.Context) error {
	return ErrNotImplemented
}
