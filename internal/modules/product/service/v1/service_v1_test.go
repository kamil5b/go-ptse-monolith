package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain/mocks"
	cachemocks "github.com/kamil5b/go-ptse-monolith/internal/shared/cache/mocks"
	eventmocks "github.com/kamil5b/go-ptse-monolith/internal/shared/events/mocks"
	uowmocks "github.com/kamil5b/go-ptse-monolith/internal/shared/uow/mocks"
)

// contextKey is a custom context key type to avoid collisions with built-in string type
type contextKey struct{}

var txContextKey = contextKey{}

// TestServiceV1_Create tests the Create method with table-driven tests
func TestServiceV1_Create(t *testing.T) {
	tests := []struct {
		name      string
		req       *domain.CreateProductRequest
		createdBy string
		repoErr   error
		wantErr   bool
		wantID    string
	}{
		{
			name: "success",
			req: &domain.CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product description",
			},
			createdBy: "user123",
			wantErr:   false,
			wantID:    "prod123",
		},
		{
			name: "repository error",
			req: &domain.CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product description",
			},
			createdBy: "user123",
			repoErr:   errors.New("database error"),
			wantErr:   true,
			wantID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
			mockEventBus := eventmocks.NewMockEventBus(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)

			ctx := context.Background()
			txCtx := context.WithValue(ctx, txContextKey, "transaction")

			mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)
			mockCache.EXPECT().GetBytes(gomock.Any(), gomock.Any()).Return(nil, errors.New("cache miss")).AnyTimes()
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			if tt.repoErr != nil {
				mockRepo.EXPECT().Create(txCtx, gomock.Any()).Return(tt.repoErr).Times(1)
				mockUOW.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Return(nil).Times(1)
			} else {
				mockRepo.EXPECT().Create(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, p *domain.Product) error {
					assert.Equal(t, tt.req.Name, p.Name)
					assert.Equal(t, tt.req.Description, p.Description)
					assert.Equal(t, tt.createdBy, p.CreatedBy)
					p.ID = tt.wantID
					return nil
				}).Times(1)
				mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
				mockUOW.EXPECT().DeferErrorContext(txCtx, nil).Return(nil).Times(1)
			}

			service := NewServiceV1(mockRepo, mockUOW, mockEventBus, mockCache)
			product, err := service.Create(ctx, tt.req, tt.createdBy)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, product)
			} else {
				require.NoError(t, err)
				require.NotNil(t, product)
				assert.Equal(t, tt.wantID, product.ID)
				assert.Equal(t, tt.req.Name, product.Name)
				assert.Equal(t, tt.req.Description, product.Description)
				assert.Equal(t, tt.createdBy, product.CreatedBy)
			}
		})
	}
}

// Test Get method with table-driven tests
func TestServiceV1_Get(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		repoErr error
		want    *domain.Product
		wantErr bool
	}{
		{
			name: "success",
			id:   "prod123",
			want: &domain.Product{
				ID:          "prod123",
				Name:        "Test Product",
				Description: "A test product",
				CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedBy:   "user123",
			},
			wantErr: false,
		},
		{
			name:    "not found",
			id:      "nonexistent",
			repoErr: errors.New("product not found"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)
			ctx := context.Background()

			mockCache.EXPECT().GetBytes(ctx, gomock.Any()).Return(nil, errors.New("cache miss")).AnyTimes()
			mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mockRepo.EXPECT().GetByID(ctx, tt.id).Return(tt.want, tt.repoErr).Times(1)

			service := NewServiceV1(mockRepo, mockUOW, nil, mockCache)
			product, err := service.Get(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, product)
			} else {
				require.NoError(t, err)
				require.NotNil(t, product)
				assert.Equal(t, tt.want.ID, product.ID)
				assert.Equal(t, tt.want.Name, product.Name)
			}
		})
	}
}

// TestServiceV1_List tests the List method with table-driven tests
func TestServiceV1_List(t *testing.T) {
	tests := []struct {
		name    string
		want    []domain.Product
		repoErr error
		wantErr bool
	}{
		{
			name: "success with products",
			want: []domain.Product{
				{
					ID:          "prod1",
					Name:        "Product 1",
					Description: "Description 1",
					CreatedAt:   time.Now(),
					CreatedBy:   "user123",
				},
				{
					ID:          "prod2",
					Name:        "Product 2",
					Description: "Description 2",
					CreatedAt:   time.Now(),
					CreatedBy:   "user456",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty list",
			want:    []domain.Product{},
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
			mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)
			ctx := context.Background()

			mockCache.EXPECT().GetBytes(ctx, gomock.Any()).Return(nil, errors.New("cache miss")).AnyTimes()
			mockRepo.EXPECT().List(ctx).Return(tt.want, tt.repoErr).Times(1)

			service := NewServiceV1(mockRepo, mockUOW, nil, mockCache)
			products, err := service.List(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, products)
			} else {
				require.NoError(t, err)
				require.NotNil(t, products)
				assert.Equal(t, len(tt.want), len(products))
				if len(tt.want) > 0 {
					assert.Equal(t, tt.want, products)
				}
			}
		})
	}
}

func TestServiceV1_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
	mockEventBus := eventmocks.NewMockEventBus(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)

	req := &domain.UpdateProductRequest{
		ID:          "prod123",
		Name:        "Updated Product",
		Description: "Updated description",
	}
	updatedBy := "user456"

	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")

	existingProduct := &domain.Product{
		ID:          "prod123",
		Name:        "Old Name",
		Description: "Old description",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		CreatedBy:   "user123",
	}

	// Expectations
	mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)
	mockRepo.EXPECT().GetByID(txCtx, req.ID).Return(existingProduct, nil).Times(1)
	mockRepo.EXPECT().Update(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, p *domain.Product) error {
		assert.Equal(t, req.Name, p.Name)
		assert.Equal(t, req.Description, p.Description)
		assert.Equal(t, updatedBy, *p.UpdatedBy)
		assert.NotNil(t, p.UpdatedAt)
		return nil
	}).Times(1)
	mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, e interface{}) error {
		event, ok := e.(domain.ProductUpdatedEvent)
		assert.True(t, ok)
		assert.Equal(t, "Updated Product", event.Name)
		assert.Equal(t, "Updated description", event.Description)
		assert.Equal(t, updatedBy, event.UpdatedBy)
		return nil
	}).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockUOW.EXPECT().DeferErrorContext(txCtx, nil).Return(nil).Times(1)

	service := NewServiceV1(mockRepo, mockUOW, mockEventBus, mockCache)
	product, err := service.Update(ctx, req, updatedBy)

	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, req.Name, product.Name)
	assert.Equal(t, req.Description, product.Description)
	assert.Equal(t, updatedBy, *product.UpdatedBy)
	assert.NotNil(t, product.UpdatedAt)
}

func TestServiceV1_Update_PartialFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
	mockEventBus := eventmocks.NewMockEventBus(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)

	req := &domain.UpdateProductRequest{
		ID:          "prod123",
		Name:        "Updated Name",
		Description: "", // Empty, should not update
	}
	updatedBy := "user456"

	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")

	existingProduct := &domain.Product{
		ID:          "prod123",
		Name:        "Old Name",
		Description: "Original description",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		CreatedBy:   "user123",
	}

	// Expectations
	mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)
	mockRepo.EXPECT().GetByID(txCtx, req.ID).Return(existingProduct, nil).Times(1)
	mockRepo.EXPECT().Update(txCtx, gomock.Any()).DoAndReturn(func(c context.Context, p *domain.Product) error {
		assert.Equal(t, "Updated Name", p.Name)
		assert.Equal(t, "Original description", p.Description) // Should remain unchanged
		return nil
	}).Times(1)
	mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
	mockCache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockUOW.EXPECT().DeferErrorContext(txCtx, nil).Return(nil).Times(1)

	service := NewServiceV1(mockRepo, mockUOW, mockEventBus, mockCache)
	product, err := service.Update(ctx, req, updatedBy)

	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, "Updated Name", product.Name)
	assert.Equal(t, "Original description", product.Description)
}

func TestServiceV1_Update_ProductNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)

	req := &domain.UpdateProductRequest{
		ID:          "nonexistent",
		Name:        "Updated Name",
		Description: "Updated description",
	}
	updatedBy := "user456"

	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")
	expectedErr := errors.New("product not found")

	// Expectations
	mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)
	mockRepo.EXPECT().GetByID(txCtx, req.ID).Return(nil, expectedErr).Times(1)
	mockUOW.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Return(nil).Times(1)

	service := NewServiceV1(mockRepo, mockUOW, nil, mockCache)
	product, err := service.Update(ctx, req, updatedBy)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, product)
}

func TestServiceV1_Update_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)

	req := &domain.UpdateProductRequest{
		ID:          "prod123",
		Name:        "Updated Product",
		Description: "Updated description",
	}
	updatedBy := "user456"

	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")

	existingProduct := &domain.Product{
		ID:          "prod123",
		Name:        "Old Name",
		Description: "Old description",
		CreatedAt:   time.Now(),
		CreatedBy:   "user123",
	}

	expectedErr := errors.New("update failed")

	// Expectations
	mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)
	mockRepo.EXPECT().GetByID(txCtx, req.ID).Return(existingProduct, nil).Times(1)
	mockRepo.EXPECT().Update(txCtx, gomock.Any()).Return(expectedErr).Times(1)
	mockUOW.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Return(nil).Times(1)

	service := NewServiceV1(mockRepo, mockUOW, nil, mockCache)
	product, err := service.Update(ctx, req, updatedBy)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, product)
}

// TestServiceV1_Delete tests the Delete method with table-driven tests
func TestServiceV1_Delete(t *testing.T) {
	type args struct {
		id        string
		deletedBy string
	}
	tests := []struct {
		name    string
		args    args
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				id:        "prod123",
				deletedBy: "user456",
			},
			wantErr: false,
		},
		{
			name: "repository error",
			args: args{
				id:        "prod123",
				deletedBy: "user456",
			},
			repoErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
			mockEventBus := eventmocks.NewMockEventBus(ctrl)
			mockCache := cachemocks.NewMockCache(ctrl)

			ctx := context.Background()
			txCtx := context.WithValue(ctx, txContextKey, "transaction")

			mockUOW.EXPECT().StartContext(ctx).Return(txCtx).Times(1)

			if tt.repoErr != nil {
				mockRepo.EXPECT().SoftDelete(txCtx, tt.args.id, tt.args.deletedBy).Return(tt.repoErr).Times(1)
				mockUOW.EXPECT().DeferErrorContext(txCtx, gomock.Any()).Return(nil).Times(1)
			} else {
				mockRepo.EXPECT().SoftDelete(txCtx, tt.args.id, tt.args.deletedBy).Return(nil).Times(1)
				mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Times(1)
				mockCache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockUOW.EXPECT().DeferErrorContext(txCtx, nil).Return(nil).Times(1)
			}

			service := NewServiceV1(mockRepo, mockUOW, mockEventBus, mockCache)
			err := service.Delete(ctx, tt.args.id, tt.args.deletedBy)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkServiceV1_Create(b *testing.B) {
	ctrl := gomock.NewController(&testing.T{})
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockUOW := uowmocks.NewMockUnitOfWork(ctrl)
	mockEventBus := eventmocks.NewMockEventBus(ctrl)
	mockCache := cachemocks.NewMockCache(ctrl)

	req := &domain.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product description",
	}
	createdBy := "user123"
	ctx := context.Background()
	txCtx := context.WithValue(ctx, txContextKey, "transaction")

	mockUOW.EXPECT().StartContext(ctx).Return(txCtx).AnyTimes()
	mockRepo.EXPECT().Create(txCtx, gomock.Any()).Return(nil).AnyTimes()
	mockEventBus.EXPECT().Publish(txCtx, gomock.Any()).Return(nil).AnyTimes()
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockUOW.EXPECT().DeferErrorContext(txCtx, nil).Return(nil).AnyTimes()

	service := NewServiceV1(mockRepo, mockUOW, mockEventBus, mockCache)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Create(ctx, req, createdBy)
	}
}
