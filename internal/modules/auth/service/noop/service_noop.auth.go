package noop

import (
	"context"
	"errors"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/auth/domain"
)

var ErrNotImplemented = errors.New("auth service not implemented")

type NoopService struct{}

func NewNoopService() *NoopService {
	return &NoopService{}
}

func (s *NoopService) Login(ctx context.Context, req *domain.LoginRequest, userAgent, ipAddress string) (*domain.LoginResponse, error) {
	return nil, ErrNotImplemented
}

func (s *NoopService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	return nil, ErrNotImplemented
}

func (s *NoopService) Logout(ctx context.Context, userID string, req *domain.LogoutRequest) error {
	return ErrNotImplemented
}

func (s *NoopService) RefreshToken(ctx context.Context, refreshToken string) (*domain.RefreshTokenResponse, error) {
	return nil, ErrNotImplemented
}

func (s *NoopService) ValidateToken(ctx context.Context, token string) (*domain.ValidateTokenResponse, error) {
	return nil, ErrNotImplemented
}

func (s *NoopService) ChangePassword(ctx context.Context, userID string, req *domain.ChangePasswordRequest) error {
	return ErrNotImplemented
}

func (s *NoopService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	return ErrNotImplemented
}

func (s *NoopService) ConfirmResetPassword(ctx context.Context, req *domain.ConfirmResetPasswordRequest) error {
	return ErrNotImplemented
}

func (s *NoopService) GetSessions(ctx context.Context, userID string) (*domain.SessionListResponse, error) {
	return nil, ErrNotImplemented
}

func (s *NoopService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	return ErrNotImplemented
}

func (s *NoopService) RevokeAllSessions(ctx context.Context, userID string) error {
	return ErrNotImplemented
}

func (s *NoopService) GenerateAccessToken(claims *domain.TokenClaims) (string, error) {
	return "", ErrNotImplemented
}

func (s *NoopService) GenerateRefreshToken(userID string) (string, error) {
	return "", ErrNotImplemented
}

func (s *NoopService) ParseToken(token string) (*domain.TokenClaims, error) {
	return nil, ErrNotImplemented
}

func (s *NoopService) HashPassword(password string) (string, error) {
	return "", ErrNotImplemented
}

func (s *NoopService) VerifyPassword(hashedPassword, password string) error {
	return ErrNotImplemented
}
