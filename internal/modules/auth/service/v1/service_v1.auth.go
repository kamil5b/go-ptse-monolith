package v1

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"go-modular-monolith/internal/modules/auth/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
	ErrPasswordMismatch   = errors.New("current password is incorrect")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthConfig struct {
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	SessionDuration      time.Duration
	BcryptCost           int
}

func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:            "supersecretkey",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		SessionDuration:      24 * time.Hour,
		BcryptCost:           bcrypt.DefaultCost,
	}
}

type ServiceV1 struct {
	repo        domain.Repository
	userCreator domain.UserCreator // ACL interface instead of direct user repo
	config      AuthConfig
}

func NewServiceV1(repo domain.Repository, userCreator domain.UserCreator, config AuthConfig) *ServiceV1 {
	return &ServiceV1{
		repo:        repo,
		userCreator: userCreator,
		config:      config,
	}
}

func (s *ServiceV1) Login(ctx context.Context, req *domain.LoginRequest, userAgent, ipAddress string) (*domain.LoginResponse, error) {
	cred, err := s.repo.GetCredentialByUsername(ctx, req.Username)
	if err != nil {
		cred, err = s.repo.GetCredentialByEmail(ctx, req.Username)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	if !cred.IsActive {
		return nil, ErrUserNotActive
	}

	if err := s.VerifyPassword(cred.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.repo.UpdateLastLogin(ctx, cred.UserID)

	claims := &domain.TokenClaims{
		UserID:   cred.UserID,
		Username: cred.Username,
		Email:    cred.Email,
	}

	accessToken, err := s.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(cred.UserID)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        uuid.NewString(),
		UserID:    cred.UserID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(s.config.RefreshTokenDuration),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(s.config.AccessTokenDuration)
	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenDuration.Seconds()),
		ExpiresAt:    expiresAt,
		User: &domain.UserInfo{
			ID:       cred.UserID,
			Username: cred.Username,
			Email:    cred.Email,
		},
	}, nil
}

func (s *ServiceV1) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	if _, err := s.repo.GetCredentialByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameExists
	}

	if _, err := s.repo.GetCredentialByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	userID := uuid.NewString()

	// Use ACL to create user - auth module doesn't depend on user module internals
	newUser := &domain.NewUser{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedBy: userID,
	}

	ctx = s.repo.StartContext(ctx)
	if err := s.userCreator.CreateUser(ctx, newUser); err != nil {
		s.repo.DeferErrorContext(ctx, err)
		return nil, err
	}

	cred := &domain.Credential{
		ID:           uuid.NewString(),
		UserID:       userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	if err := s.repo.CreateCredential(ctx, cred); err != nil {
		s.repo.DeferErrorContext(ctx, err)
		return nil, err
	}

	s.repo.DeferErrorContext(ctx, nil)

	return &domain.RegisterResponse{
		User: &domain.UserInfo{
			ID:       userID,
			Username: req.Username,
			Email:    req.Email,
			Name:     req.Name,
		},
		Message: "Registration successful",
	}, nil
}

func (s *ServiceV1) Logout(ctx context.Context, userID string, req *domain.LogoutRequest) error {
	if req.AllDevices {
		return s.repo.RevokeAllUserSessions(ctx, userID)
	}

	if req.RefreshToken != "" {
		session, err := s.repo.GetSessionByToken(ctx, req.RefreshToken)
		if err != nil {
			return ErrSessionNotFound
		}
		return s.repo.RevokeSession(ctx, session.ID)
	}

	return nil
}

func (s *ServiceV1) RefreshToken(ctx context.Context, refreshToken string) (*domain.RefreshTokenResponse, error) {
	session, err := s.repo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	cred, err := s.repo.GetCredentialByUserID(ctx, session.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !cred.IsActive {
		return nil, ErrUserNotActive
	}

	claims := &domain.TokenClaims{
		UserID:   cred.UserID,
		Username: cred.Username,
		Email:    cred.Email,
	}

	accessToken, err := s.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(s.config.AccessTokenDuration)
	return &domain.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenDuration.Seconds()),
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *ServiceV1) ValidateToken(ctx context.Context, token string) (*domain.ValidateTokenResponse, error) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return &domain.ValidateTokenResponse{
			Valid:  false,
			Reason: err.Error(),
		}, nil
	}

	cred, err := s.repo.GetCredentialByUserID(ctx, claims.UserID)
	if err != nil {
		return &domain.ValidateTokenResponse{
			Valid:  false,
			Reason: "user not found",
		}, nil
	}

	if !cred.IsActive {
		return &domain.ValidateTokenResponse{
			Valid:  false,
			Reason: "user account is inactive",
		}, nil
	}

	return &domain.ValidateTokenResponse{
		Valid: true,
		User: &domain.UserInfo{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    claims.Roles,
		},
	}, nil
}

func (s *ServiceV1) ChangePassword(ctx context.Context, userID string, req *domain.ChangePasswordRequest) error {
	cred, err := s.repo.GetCredentialByUserID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.VerifyPassword(cred.PasswordHash, req.OldPassword); err != nil {
		return ErrPasswordMismatch
	}

	hashedPassword, err := s.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, hashedPassword)
}

func (s *ServiceV1) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	return nil
}

func (s *ServiceV1) ConfirmResetPassword(ctx context.Context, req *domain.ConfirmResetPasswordRequest) error {
	return nil
}

func (s *ServiceV1) GetSessions(ctx context.Context, userID string) (*domain.SessionListResponse, error) {
	sessions, err := s.repo.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessionInfos := make([]domain.SessionInfo, len(sessions))
	for i, sess := range sessions {
		sessionInfos[i] = domain.SessionInfo{
			ID:        sess.ID,
			UserAgent: sess.UserAgent,
			IPAddress: sess.IPAddress,
			CreatedAt: sess.CreatedAt,
			ExpiresAt: sess.ExpiresAt,
		}
	}

	return &domain.SessionListResponse{Sessions: sessionInfos}, nil
}

func (s *ServiceV1) RevokeSession(ctx context.Context, userID, sessionID string) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}

	if session.UserID != userID {
		return ErrSessionNotFound
	}

	return s.repo.RevokeSession(ctx, sessionID)
}

func (s *ServiceV1) RevokeAllSessions(ctx context.Context, userID string) error {
	return s.repo.RevokeAllUserSessions(ctx, userID)
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles,omitempty"`
}

func (s *ServiceV1) GenerateAccessToken(claims *domain.TokenClaims) (string, error) {
	now := time.Now().UTC()
	jwtClaims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *ServiceV1) GenerateRefreshToken(userID string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *ServiceV1) ParseToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return &domain.TokenClaims{
			UserID:   claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    claims.Roles,
		}, nil
	}

	return nil, ErrInvalidToken
}

func (s *ServiceV1) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
	return string(bytes), err
}

func (s *ServiceV1) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
