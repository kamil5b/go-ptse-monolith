package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	productDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	productv1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/proto"
)

// MockProductService is a mock implementation of productDomain.Service
type MockProductService struct {
	CreateFunc func(ctx context.Context, req *productDomain.CreateProductRequest, createdBy string) (*productDomain.Product, error)
	GetFunc    func(ctx context.Context, id string) (*productDomain.Product, error)
	ListFunc   func(ctx context.Context) ([]productDomain.Product, error)
	UpdateFunc func(ctx context.Context, req *productDomain.UpdateProductRequest, updatedBy string) (*productDomain.Product, error)
	DeleteFunc func(ctx context.Context, id, deletedBy string) error
}

func (m *MockProductService) Create(ctx context.Context, req *productDomain.CreateProductRequest, createdBy string) (*productDomain.Product, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req, createdBy)
	}
	return nil, errors.New("not implemented")
}

func (m *MockProductService) Get(ctx context.Context, id string) (*productDomain.Product, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockProductService) List(ctx context.Context) ([]productDomain.Product, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockProductService) Update(ctx context.Context, req *productDomain.UpdateProductRequest, updatedBy string) (*productDomain.Product, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, req, updatedBy)
	}
	return nil, errors.New("not implemented")
}

func (m *MockProductService) Delete(ctx context.Context, id, deletedBy string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id, deletedBy)
	}
	return errors.New("not implemented")
}

// TestGRPCHandler_Create tests the Create method
func TestGRPCHandler_Create(t *testing.T) {
	mockService := &MockProductService{
		CreateFunc: func(ctx context.Context, req *productDomain.CreateProductRequest, createdBy string) (*productDomain.Product, error) {
			return &productDomain.Product{
				ID:          "product-1",
				Name:        req.Name,
				Description: req.Description,
				CreatedAt:   time.Now(),
				CreatedBy:   createdBy,
			}, nil
		},
	}

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
	now := time.Now()

	mockService := &MockProductService{
		GetFunc: func(ctx context.Context, id string) (*productDomain.Product, error) {
			if id == "product-1" {
				return &productDomain.Product{
					ID:          id,
					Name:        "Test Product",
					Description: "A test product",
					CreatedAt:   now,
					CreatedBy:   "user-123",
				}, nil
			}
			return nil, errors.New("product not found")
		},
	}

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
	now := time.Now()

	mockService := &MockProductService{
		ListFunc: func(ctx context.Context) ([]productDomain.Product, error) {
			return []productDomain.Product{
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
			}, nil
		},
	}

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
	now := time.Now()
	updatedName := "Updated Product"

	mockService := &MockProductService{
		UpdateFunc: func(ctx context.Context, req *productDomain.UpdateProductRequest, updatedBy string) (*productDomain.Product, error) {
			return &productDomain.Product{
				ID:          req.ID,
				Name:        updatedName,
				Description: req.Description,
				CreatedAt:   now,
				CreatedBy:   "user-123",
				UpdatedAt:   &now,
				UpdatedBy:   &updatedBy,
			}, nil
		},
	}

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
	mockService := &MockProductService{
		DeleteFunc: func(ctx context.Context, id, deletedBy string) error {
			if id == "product-1" {
				return nil
			}
			return errors.New("product not found")
		},
	}

	handler := NewGRPCHandler(mockService)
	ctx := context.WithValue(context.Background(), "user_id", "user-123")

	_, err := handler.Delete(ctx, &productv1.DeleteProductRequest{Id: "product-1"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
