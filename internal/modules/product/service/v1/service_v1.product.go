package v1

import (
	"context"
	"go-modular-monolith/internal/domain/product"
	"go-modular-monolith/internal/domain/uow"
	"time"
)

type ServiceV1 struct {
	repo product.ProductRepository
	uow  uow.UnitOfWork
}

func NewServiceV1(r product.ProductRepository, u uow.UnitOfWork) *ServiceV1 {
	return &ServiceV1{repo: r, uow: u}
}

func (s *ServiceV1) Create(ctx context.Context, req *product.CreateProductRequest, createdBy string) (*product.Product, error) {
	ctx = s.uow.StartContext(ctx)
	var p product.Product
	p.Name = req.Name
	p.Description = req.Description
	p.CreatedAt = time.Now().UTC()
	p.CreatedBy = createdBy
	err := s.repo.Create(ctx, &p)
	if err != nil {
		s.uow.DeferErrorContext(ctx, err)
		return nil, err
	}
	return &p, nil
}
func (s *ServiceV1) Get(ctx context.Context, id string) (*product.Product, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *ServiceV1) List(ctx context.Context) ([]product.Product, error) {
	return s.repo.List(ctx)
}
func (s *ServiceV1) Update(ctx context.Context, req *product.UpdateProductRequest, updatedBy string) (*product.Product, error) {
	p, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	now := time.Now().UTC()
	p.UpdatedAt = &now
	p.UpdatedBy = &updatedBy
	err = s.repo.Update(ctx, p)
	if err != nil {
		s.uow.DeferErrorContext(ctx, err)
		return nil, err
	}
	return p, nil
}
func (s *ServiceV1) Delete(ctx context.Context, id, by string) error {
	return s.repo.SoftDelete(ctx, id, by)
}
