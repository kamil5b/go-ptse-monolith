package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain/mocks"
	cachemocks "github.com/kamil5b/go-ptse-monolith/internal/shared/cache/mocks"
	emailmocks "github.com/kamil5b/go-ptse-monolith/internal/shared/email/mocks"
	eventmocks "github.com/kamil5b/go-ptse-monolith/internal/shared/events/mocks"
)

// contextKey is a custom context key type
type contextKey struct{}

var txContextKey = contextKey{}

// TestServiceV1_Create tests the Create method with table-driven tests
func TestServiceV1_Create(t *testing.T) {
	tests := []struct {
		name      string
		req       *domain.CreateUserRequest
		createdBy string
		repoErr   error
		wantErr   bool
		wantID    string
	}{
		{
			name: "success",
			req: &domain.CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			createdBy: "admin123",
			wantErr:   false,
			wantID:    "user123",
		},
		{
			name: "repository error",
			req: &domain.CreateUserRequest{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
			createdBy: "admin123",
			repoErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockEventBus := eventmocks.NewMockEventBus(ctrl)
			mockEmailService := emailmocks.NewMockEmailService(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)

			ctx := context.Background()
			txCtx := context.WithValue(ctx, txContextKey, "transaction")

			mockRepo.EXPECT().StartContext(ctx).Return(txCtx).Times(1)

			if tt.repoErr != nil {
				mockRepo.EXPECT().Create(txCtx, gomock.Any()).Return(tt.repoErr).Times(1)
				mockRepo.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Times(1)
			} else {
				mockRepo.EXPECT().Create(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, u *domain.User) error {
					assert.Equal(t, tt.req.Name, u.Name)
					assert.Equal(t, tt.req.Email, u.Email)
					assert.Equal(t, tt.createdBy, u.CreatedBy)
					u.ID = tt.wantID
					return nil
				}).Times(1)
				mockEmailService.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
				mockRepo.EXPECT().DeferErrorContext(txCtx, nil).Times(1)
			}

			service := NewServiceV1(mockRepo, mockEventBus, mockEmailService, mockCache)
			user, err := service.Create(ctx, tt.req, tt.createdBy)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.wantID, user.ID)
				assert.Equal(t, tt.req.Name, user.Name)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Equal(t, tt.createdBy, user.CreatedBy)
			}
		})
	}
}

// TestServiceV1_Get tests the Get method
func TestServiceV1_Get(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		repoErr error
		want    *domain.User
		wantErr bool
	}{
		{
			name: "success",
			id:   "user123",
			want: &domain.User{
				ID:        "user123",
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedBy: "admin",
			},
			wantErr: false,
		},
		{
			name:    "not found",
			id:      "nonexistent",
			repoErr: errors.New("user not found"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)
			ctx := context.Background()

			// Cache miss
			mockCache.EXPECT().GetBytes(ctx, gomock.Any()).Return(nil, errors.New("cache miss")).Times(1)

			if tt.wantErr {
				mockRepo.EXPECT().GetByID(ctx, tt.id).Return(nil, tt.repoErr).Times(1)
			} else {
				mockRepo.EXPECT().GetByID(ctx, tt.id).Return(tt.want, nil).Times(1)
				mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			}

			service := NewServiceV1(mockRepo, nil, nil, mockCache)
			user, err := service.Get(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.want.ID, user.ID)
				assert.Equal(t, tt.want.Name, user.Name)
				assert.Equal(t, tt.want.Email, user.Email)
			}
		})
	}
}

// TestServiceV1_List tests the List method
func TestServiceV1_List(t *testing.T) {
	tests := []struct {
		name    string
		want    []domain.User
		repoErr error
		wantErr bool
	}{
		{
			name: "success with users",
			want: []domain.User{
				{
					ID:        "user1",
					Name:      "User 1",
					Email:     "user1@example.com",
					CreatedAt: time.Now(),
					CreatedBy: "admin",
				},
				{
					ID:        "user2",
					Name:      "User 2",
					Email:     "user2@example.com",
					CreatedAt: time.Now(),
					CreatedBy: "admin",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty list",
			want:    []domain.User{},
			wantErr: false,
		},
		{
			name:    "database error",
			want:    nil,
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			ctx := context.Background()

			mockRepo.EXPECT().List(ctx).Return(tt.want, tt.repoErr).Times(1)

			service := NewServiceV1(mockRepo, nil, nil, nil)
			users, err := service.List(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				require.NoError(t, err)
				require.NotNil(t, users)
				assert.Equal(t, len(tt.want), len(users))
			}
		})
	}
}

// TestServiceV1_Update tests the Update method
func TestServiceV1_Update(t *testing.T) {
	tests := []struct {
		name      string
		req       *domain.UpdateUserRequest
		updatedBy string
		repoErr   error
		wantErr   bool
	}{
		{
			name: "success - full update",
			req: &domain.UpdateUserRequest{
				ID:    "user123",
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			updatedBy: "admin456",
			wantErr:   false,
		},
		{
			name: "success - partial update (name only)",
			req: &domain.UpdateUserRequest{
				ID:   "user123",
				Name: "Updated Name",
			},
			updatedBy: "admin456",
			wantErr:   false,
		},
		{
			name: "user not found",
			req: &domain.UpdateUserRequest{
				ID:   "nonexistent",
				Name: "Updated Name",
			},
			updatedBy: "admin456",
			repoErr:   errors.New("user not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockEventBus := eventmocks.NewMockEventBus(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)

			ctx := context.Background()
			txCtx := context.WithValue(ctx, txContextKey, "transaction")

			existingUser := &domain.User{
				ID:        tt.req.ID,
				Name:      "Old Name",
				Email:     "old@example.com",
				CreatedAt: time.Now().Add(-24 * time.Hour),
				CreatedBy: "admin",
			}

			mockRepo.EXPECT().StartContext(ctx).Return(txCtx).Times(1)

			if tt.repoErr != nil {
				mockRepo.EXPECT().GetByID(txCtx, tt.req.ID).Return(nil, tt.repoErr).Times(1)
				mockRepo.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Times(1)
			} else {
				mockRepo.EXPECT().GetByID(txCtx, tt.req.ID).Return(existingUser, nil).Times(1)
				mockRepo.EXPECT().Update(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, u *domain.User) error {
					if tt.req.Name != "" {
						assert.Equal(t, tt.req.Name, u.Name)
					}
					if tt.req.Email != "" {
						assert.Equal(t, tt.req.Email, u.Email)
					}
					assert.Equal(t, tt.updatedBy, *u.UpdatedBy)
					assert.NotNil(t, u.UpdatedAt)
					return nil
				}).Times(1)
				mockCache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
				mockRepo.EXPECT().DeferErrorContext(txCtx, nil).Times(1)
			}

			service := NewServiceV1(mockRepo, mockEventBus, nil, mockCache)
			user, err := service.Update(ctx, tt.req, tt.updatedBy)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				if tt.req.Name != "" {
					assert.Equal(t, tt.req.Name, user.Name)
				}
				if tt.req.Email != "" {
					assert.Equal(t, tt.req.Email, user.Email)
				}
			}
		})
	}
}

// TestServiceV1_Delete tests the Delete method
func TestServiceV1_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		deletedBy string
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "success",
			id:        "user123",
			deletedBy: "admin456",
			wantErr:   false,
		},
		{
			name:      "repository error",
			id:        "user123",
			deletedBy: "admin456",
			repoErr:   errors.New("delete failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockEventBus := eventmocks.NewMockEventBus(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)

			ctx := context.Background()
			txCtx := context.WithValue(ctx, txContextKey, "transaction")

			mockRepo.EXPECT().StartContext(ctx).Return(txCtx).Times(1)

			if tt.repoErr != nil {
				mockRepo.EXPECT().SoftDelete(txCtx, tt.id, tt.deletedBy).Return(tt.repoErr).Times(1)
				mockRepo.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Times(1)
			} else {
				mockRepo.EXPECT().SoftDelete(txCtx, tt.id, tt.deletedBy).Return(nil).Times(1)
				mockCache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
				mockRepo.EXPECT().DeferErrorContext(txCtx, nil).Times(1)
			}

			service := NewServiceV1(mockRepo, mockEventBus, nil, mockCache)
			err := service.Delete(ctx, tt.id, tt.deletedBy)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestServiceV1_GetWithCache tests cache hit scenario
func TestServiceV1_GetWithCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)
	ctx := context.Background()

	cachedUser := &domain.User{
		ID:        "user123",
		Name:      "Cached User",
		Email:     "cached@example.com",
		CreatedAt: time.Now(),
		CreatedBy: "admin",
	}

	// Mock cache hit
	mockCache.EXPECT().GetBytes(ctx, gomock.Any()).DoAndReturn(func(c context.Context, key string) ([]byte, error) {
		return []byte(`{"ID":"user123","Name":"Cached User","Email":"cached@example.com"}`), nil
	}).Times(1)

	// Repository should not be called on cache hit
	mockRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(0)

	service := NewServiceV1(mockRepo, nil, nil, mockCache)
	user, err := service.Get(ctx, "user123")

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, cachedUser.ID, user.ID)
	assert.Equal(t, cachedUser.Name, user.Name)
	assert.Equal(t, cachedUser.Email, user.Email)
}

// Benchmark tests
func BenchmarkServiceV1_Create(b *testing.B) {
	ctrl := gomock.NewController(&testing.T{})
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockEventBus := eventmocks.NewMockEventBus(ctrl)
	mockEmailService := emailmocks.NewMockEmailService(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)

	req := &domain.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	createdBy := "admin123"
	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")

	mockRepo.EXPECT().StartContext(ctx).Return(txCtx).AnyTimes()
	mockRepo.EXPECT().Create(txCtx, gomock.Any()).Return(nil).AnyTimes()
	mockEmailService.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().DeferErrorContext(txCtx, nil).AnyTimes()

	service := NewServiceV1(mockRepo, mockEventBus, mockEmailService, mockCache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Create(ctx, req, createdBy)
	}
}
