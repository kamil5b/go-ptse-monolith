package domain

import (
	"context"

	sharedctx "go-modular-monolith/internal/shared/context"
)

// Handler defines the interface for product HTTP handlers
type Handler interface {
	Create(c sharedctx.Context) error
	Get(c sharedctx.Context) error
	List(c sharedctx.Context) error
	Update(c sharedctx.Context) error
	Delete(c sharedctx.Context) error
}

// Service defines the interface for product business logic
type Service interface {
	Create(ctx context.Context, req *CreateProductRequest, createdBy string) (*Product, error)
	Get(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]Product, error)
	Update(ctx context.Context, req *UpdateProductRequest, updatedBy string) (*Product, error)
	Delete(ctx context.Context, id, deletedBy string) error
}

// Repository defines the interface for product data access
type Repository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]Product, error)
	Update(ctx context.Context, p *Product) error
	SoftDelete(ctx context.Context, id, deletedBy string) error
}
