package product

import (
	"context"
)

type Context interface {
	BindJSON(obj any) error
	BindURI(obj any) error
	BindQuery(obj any) error
	BindHeader(obj any) error
	Bind(obj any) error
	JSON(code int, v any) error
	Param(name string) string
	GetUserID() string
	Get(key string) any
	GetContext() context.Context
}

type ProductHandler interface {
	Create(c Context) error
	Get(c Context) error
	List(c Context) error
	Update(c Context) error
	Delete(c Context) error
}

type ProductService interface {
	Create(ctx context.Context, req *CreateProductRequest, createdBy string) (*Product, error)
	Get(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]Product, error)
	Update(ctx context.Context, req *UpdateProductRequest, updatedBy string) (*Product, error)
	Delete(ctx context.Context, id, deletedBy string) error
}

type ProductRepository interface {
	StartContext(ctx context.Context) context.Context
	DeferErrorContext(ctx context.Context, err error)

	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]Product, error)
	Update(ctx context.Context, p *Product) error
	SoftDelete(ctx context.Context, id, deletedBy string) error
}
