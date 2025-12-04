package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/auth/domain"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/auth/domain/mocks"
)

// contextKey is a custom context key type
type contextKey struct{}

var txContextKey = contextKey{}

func TestServiceV1_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	hashedPassword, _ := service.HashPassword("password123")

	cred := &domain.Credential{
		ID:           "cred123",
		UserID:       "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	req := &domain.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.EXPECT().GetCredentialByUsername(ctx, req.Username).Return(cred, nil).Times(1)
	mockRepo.EXPECT().UpdateLastLogin(ctx, cred.UserID).Return(nil).Times(1)
	mockRepo.EXPECT().CreateSession(ctx, gomock.Any()).Return(nil).Times(1)

	resp, err := service.Login(ctx, req, "Mozilla/5.0", "192.168.1.1")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, cred.UserID, resp.User.ID)
	assert.Equal(t, cred.Username, resp.User.Username)
	assert.Equal(t, cred.Email, resp.User.Email)
}

func TestServiceV1_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	req := &domain.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockRepo.EXPECT().GetCredentialByUsername(ctx, req.Username).Return(nil, errors.New("not found")).Times(1)
	mockRepo.EXPECT().GetCredentialByEmail(ctx, req.Username).Return(nil, errors.New("not found")).Times(1)

	resp, err := service.Login(ctx, req, "Mozilla/5.0", "192.168.1.1")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
}

func TestServiceV1_Login_InactiveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	hashedPassword, _ := service.HashPassword("password123")

	cred := &domain.Credential{
		ID:           "cred123",
		UserID:       "user123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		IsActive:     false, // Inactive user
	}

	req := &domain.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.EXPECT().GetCredentialByUsername(ctx, req.Username).Return(cred, nil).Times(1)

	resp, err := service.Login(ctx, req, "Mozilla/5.0", "192.168.1.1")

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotActive, err)
	assert.Nil(t, resp)
}

func TestServiceV1_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")

	req := &domain.RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "New User",
	}

	mockRepo.EXPECT().GetCredentialByUsername(ctx, req.Username).Return(nil, errors.New("not found")).Times(1)
	mockRepo.EXPECT().GetCredentialByEmail(ctx, req.Email).Return(nil, errors.New("not found")).Times(1)
	mockRepo.EXPECT().StartContext(ctx).Return(txCtx).Times(1)
	mockUserCreator.EXPECT().CreateUser(txCtx, gomock.Any()).Return(nil).Times(1)
	mockRepo.EXPECT().CreateCredential(txCtx, gomock.Any()).Return(nil).Times(1)
	mockRepo.EXPECT().DeferErrorContext(txCtx, nil).Times(1)

	resp, err := service.Register(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, req.Username, resp.User.Username)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.Equal(t, "Registration successful", resp.Message)
}

func TestServiceV1_Register_UsernameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	req := &domain.RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	}

	existingCred := &domain.Credential{
		ID:       "cred123",
		UserID:   "user123",
		Username: "existinguser",
	}

	mockRepo.EXPECT().GetCredentialByUsername(ctx, req.Username).Return(existingCred, nil).Times(1)

	resp, err := service.Register(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrUsernameExists, err)
	assert.Nil(t, resp)
}

func TestServiceV1_Register_EmailExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	req := &domain.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "New User",
	}

	existingCred := &domain.Credential{
		ID:     "cred123",
		UserID: "user123",
		Email:  "existing@example.com",
	}

	mockRepo.EXPECT().GetCredentialByUsername(ctx, req.Username).Return(nil, errors.New("not found")).Times(1)
	mockRepo.EXPECT().GetCredentialByEmail(ctx, req.Email).Return(existingCred, nil).Times(1)

	resp, err := service.Register(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
	assert.Nil(t, resp)
}

func TestServiceV1_Logout_SingleDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"
	refreshToken := "token123"

	req := &domain.LogoutRequest{
		RefreshToken: refreshToken,
		AllDevices:   false,
	}

	session := &domain.Session{
		ID:     "session123",
		UserID: userID,
		Token:  refreshToken,
	}

	mockRepo.EXPECT().GetSessionByToken(ctx, refreshToken).Return(session, nil).Times(1)
	mockRepo.EXPECT().RevokeSession(ctx, session.ID).Return(nil).Times(1)

	err := service.Logout(ctx, userID, req)

	require.NoError(t, err)
}

func TestServiceV1_Logout_AllDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"

	req := &domain.LogoutRequest{
		AllDevices: true,
	}

	mockRepo.EXPECT().RevokeAllUserSessions(ctx, userID).Return(nil).Times(1)

	err := service.Logout(ctx, userID, req)

	require.NoError(t, err)
}

func TestServiceV1_RefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	refreshToken := "refresh_token_123"

	session := &domain.Session{
		ID:        "session123",
		UserID:    "user123",
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	cred := &domain.Credential{
		ID:       "cred123",
		UserID:   "user123",
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}

	mockRepo.EXPECT().GetSessionByToken(ctx, refreshToken).Return(session, nil).Times(1)
	mockRepo.EXPECT().GetCredentialByUserID(ctx, session.UserID).Return(cred, nil).Times(1)

	resp, err := service.RefreshToken(ctx, refreshToken)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, refreshToken, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestServiceV1_RefreshToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	refreshToken := "invalid_token"

	mockRepo.EXPECT().GetSessionByToken(ctx, refreshToken).Return(nil, errors.New("not found")).Times(1)

	resp, err := service.RefreshToken(ctx, refreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, resp)
}

func TestServiceV1_ValidateToken_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()

	// Generate a valid token
	claims := &domain.TokenClaims{
		UserID:   "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"user"},
	}

	token, err := service.GenerateAccessToken(claims)
	require.NoError(t, err)

	cred := &domain.Credential{
		ID:       "cred123",
		UserID:   "user123",
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}

	mockRepo.EXPECT().GetCredentialByUserID(ctx, claims.UserID).Return(cred, nil).Times(1)

	resp, err := service.ValidateToken(ctx, token)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Valid)
	assert.Equal(t, claims.UserID, resp.User.ID)
	assert.Equal(t, claims.Username, resp.User.Username)
}

func TestServiceV1_ValidateToken_Invalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	invalidToken := "invalid.token.string"

	resp, err := service.ValidateToken(ctx, invalidToken)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Valid)
	assert.NotEmpty(t, resp.Reason)
}

func TestServiceV1_ChangePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"
	oldPassword := "oldpass123"
	newPassword := "newpass456"

	hashedOldPassword, _ := service.HashPassword(oldPassword)

	cred := &domain.Credential{
		ID:           "cred123",
		UserID:       userID,
		PasswordHash: hashedOldPassword,
	}

	req := &domain.ChangePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	mockRepo.EXPECT().GetCredentialByUserID(ctx, userID).Return(cred, nil).Times(1)
	mockRepo.EXPECT().UpdatePassword(ctx, userID, gomock.Any()).Return(nil).Times(1)

	err := service.ChangePassword(ctx, userID, req)

	require.NoError(t, err)
}

func TestServiceV1_ChangePassword_WrongOldPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"

	hashedPassword, _ := service.HashPassword("correctpass")

	cred := &domain.Credential{
		ID:           "cred123",
		UserID:       userID,
		PasswordHash: hashedPassword,
	}

	req := &domain.ChangePasswordRequest{
		OldPassword: "wrongpass",
		NewPassword: "newpass456",
	}

	mockRepo.EXPECT().GetCredentialByUserID(ctx, userID).Return(cred, nil).Times(1)

	err := service.ChangePassword(ctx, userID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrPasswordMismatch, err)
}

func TestServiceV1_GetSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"

	sessions := []domain.Session{
		{
			ID:        "session1",
			UserID:    userID,
			UserAgent: "Mozilla/5.0",
			IPAddress: "192.168.1.1",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:        "session2",
			UserID:    userID,
			UserAgent: "Chrome/90.0",
			IPAddress: "192.168.1.2",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	mockRepo.EXPECT().GetSessionsByUserID(ctx, userID).Return(sessions, nil).Times(1)

	resp, err := service.GetSessions(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Sessions))
	assert.Equal(t, sessions[0].ID, resp.Sessions[0].ID)
	assert.Equal(t, sessions[1].ID, resp.Sessions[1].ID)
}

func TestServiceV1_RevokeSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"
	sessionID := "session123"

	session := &domain.Session{
		ID:     sessionID,
		UserID: userID,
	}

	mockRepo.EXPECT().GetSessionByID(ctx, sessionID).Return(session, nil).Times(1)
	mockRepo.EXPECT().RevokeSession(ctx, sessionID).Return(nil).Times(1)

	err := service.RevokeSession(ctx, userID, sessionID)

	require.NoError(t, err)
}

func TestServiceV1_RevokeAllSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUserCreator := mocks.NewMockUserCreator(ctrl)

	config := DefaultAuthConfig()
	service := NewServiceV1(mockRepo, mockUserCreator, config)

	ctx := context.Background()
	userID := "user123"

	mockRepo.EXPECT().RevokeAllUserSessions(ctx, userID).Return(nil).Times(1)

	err := service.RevokeAllSessions(ctx, userID)

	require.NoError(t, err)
}

func TestServiceV1_HashAndVerifyPassword(t *testing.T) {
	config := DefaultAuthConfig()
	service := NewServiceV1(nil, nil, config)

	password := "mypassword123"

	// Test hash
	hash, err := service.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Test verify with correct password
	err = service.VerifyPassword(hash, password)
	assert.NoError(t, err)

	// Test verify with wrong password
	err = service.VerifyPassword(hash, "wrongpassword")
	assert.Error(t, err)
}

func TestServiceV1_TokenGeneration(t *testing.T) {
	config := DefaultAuthConfig()
	service := NewServiceV1(nil, nil, config)

	claims := &domain.TokenClaims{
		UserID:   "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"user", "admin"},
	}

	// Test access token generation
	accessToken, err := service.GenerateAccessToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Test token parsing
	parsedClaims, err := service.ParseToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, parsedClaims.UserID)
	assert.Equal(t, claims.Username, parsedClaims.Username)
	assert.Equal(t, claims.Email, parsedClaims.Email)
	assert.Equal(t, claims.Roles, parsedClaims.Roles)

	// Test refresh token generation
	refreshToken, err := service.GenerateRefreshToken("user123")
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
}

// Benchmark tests
func BenchmarkServiceV1_HashPassword(b *testing.B) {
	config := DefaultAuthConfig()
	service := NewServiceV1(nil, nil, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.HashPassword("password123")
	}
}

func BenchmarkServiceV1_VerifyPassword(b *testing.B) {
	config := DefaultAuthConfig()
	service := NewServiceV1(nil, nil, config)

	hash, _ := service.HashPassword("password123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.VerifyPassword(hash, "password123")
	}
}

func BenchmarkServiceV1_GenerateAccessToken(b *testing.B) {
	config := DefaultAuthConfig()
	service := NewServiceV1(nil, nil, config)

	claims := &domain.TokenClaims{
		UserID:   "user123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GenerateAccessToken(claims)
	}
}
