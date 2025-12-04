package grpc

import (
	"context"
	"testing"
	"time"

	productDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	mockdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain/mocks"
	productv1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/proto/v1"

	gomock "github.com/golang/mock/gomock"
)

// TestGRPCHandler_Create tests the Create method
func TestGRPCHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockdomain.NewMockService(ctrl)
	now := time.Now()

	mockService.EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(&productDomain.CreateProductRequest{}), "user-123").
		Return(&productDomain.Product{
			ID:          "product-1",
			Name:        "Test Product",
			Description: "A test product",
			CreatedAt:   now,
			CreatedBy:   "user-123",
		}, nil)

	handler := NewGRPCHandler(mockService)
	ctx := context.WithValue(context.Background(), "user_id", "user-123")

	resp, err := handler.Create(ctx, &productv1.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Product.Name != "Test Product" {
		t.Errorf("expected name 'Test Product', got %q", resp.Product.Name)
	}

	if resp.Product.CreatedBy != "user-123" {
		t.Errorf("expected createdBy 'user-123', got %q", resp.Product.CreatedBy)
	}
}

// TestGRPCHandler_Get tests the Get method
func TestGRPCHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockdomain.NewMockService(ctrl)
	now := time.Now()

	mockService.EXPECT().
		Get(gomock.Any(), "product-1").
		Return(&productDomain.Product{
			ID:          "product-1",
			Name:        "Test Product",
			Description: "A test product",
			CreatedAt:   now,
			CreatedBy:   "user-123",
		}, nil)

	handler := NewGRPCHandler(mockService)

	resp, err := handler.Get(context.Background(), &productv1.GetProductRequest{Id: "product-1"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Product.Id != "product-1" {
		t.Errorf("expected Id 'product-1', got %q", resp.Product.Id)
	}
}

// TestGRPCHandler_List tests the List method
func TestGRPCHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockdomain.NewMockService(ctrl)
	now := time.Now()

	mockService.EXPECT().
		List(gomock.Any()).
		Return([]productDomain.Product{
			{
				ID:          "product-1",
				Name:        "Product 1",
				Description: "First product",
				CreatedAt:   now,
				CreatedBy:   "user-123",
			},
			{
				ID:          "product-2",
				Name:        "Product 2",
				Description: "Second product",
				CreatedAt:   now,
				CreatedBy:   "user-123",
			},
		}, nil)

	handler := NewGRPCHandler(mockService)

	resp, err := handler.List(context.Background(), nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(resp.Products))
	}
}

// TestGRPCHandler_Update tests the Update method
func TestGRPCHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockdomain.NewMockService(ctrl)
	now := time.Now()
	updatedName := "Updated Product"
	updatedBy := "user-456"

	mockService.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(&productDomain.UpdateProductRequest{}), "user-456").
		Return(&productDomain.Product{
			ID:          "product-1",
			Name:        updatedName,
			Description: "",
			CreatedAt:   now,
			CreatedBy:   "user-123",
			UpdatedAt:   &now,
			UpdatedBy:   &updatedBy,
		}, nil)

	handler := NewGRPCHandler(mockService)
	ctx := context.WithValue(context.Background(), "user_id", "user-456")

	resp, err := handler.Update(ctx, &productv1.UpdateProductRequest{
		Id:   "product-1",
		Name: &updatedName,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Product.Name != updatedName {
		t.Errorf("expected name %q, got %q", updatedName, resp.Product.Name)
	}

	if resp.Product.UpdatedBy == nil || *resp.Product.UpdatedBy != "user-456" {
		t.Errorf("expected updatedBy 'user-456'")
	}
}

// TestGRPCHandler_Delete tests the Delete method
func TestGRPCHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockdomain.NewMockService(ctrl)

	mockService.EXPECT().
		Delete(gomock.Any(), "product-1", "user-123").
		Return(nil)

	handler := NewGRPCHandler(mockService)
	ctx := context.WithValue(context.Background(), "user_id", "user-123")

	_, err := handler.Delete(ctx, &productv1.DeleteProductRequest{Id: "product-1"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
