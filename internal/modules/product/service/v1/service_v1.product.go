package v1

import (
	"context"
	"fmt"
	"time"

	"go-modular-monolith/internal/modules/product/domain"
	"go-modular-monolith/internal/shared/events"
	"go-modular-monolith/internal/shared/uow"
)

type ServiceV1 struct {
	repo     domain.Repository
	uow      uow.UnitOfWork
	eventBus events.EventBus
}

func NewServiceV1(r domain.Repository, u uow.UnitOfWork, eb events.EventBus) *ServiceV1 {
	return &ServiceV1{repo: r, uow: u, eventBus: eb}
}

func (s *ServiceV1) Create(ctx context.Context, req *domain.CreateProductRequest, createdBy string) (product *domain.Product, err error) {
	ctx = s.uow.StartContext(ctx)
	defer s.uow.DeferErrorContext(ctx, err)
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("panic: %s", x)
			case error:
				err = fmt.Errorf("panic: %w", x)
			default:
				err = fmt.Errorf("panic: %v", x)
			}
		}
	}()

	var p domain.Product
	p.Name = req.Name
	p.Description = req.Description
	p.CreatedAt = time.Now().UTC()
	p.CreatedBy = createdBy

	if err = s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.ProductCreatedEvent{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			CreatedBy:   createdBy,
			CreatedAt:   p.CreatedAt,
		})
	}

	product = &p
	return
}
func (s *ServiceV1) Get(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *ServiceV1) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}
func (s *ServiceV1) Update(ctx context.Context, req *domain.UpdateProductRequest, updatedBy string) (product *domain.Product, err error) {
	ctx = s.uow.StartContext(ctx)
	defer s.uow.DeferErrorContext(ctx, err)
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("panic: %s", x)
			case error:
				err = fmt.Errorf("panic: %w", x)
			default:
				err = fmt.Errorf("panic: %v", x)
			}
		}
	}()

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
		return nil, err
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.ProductUpdatedEvent{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			UpdatedBy:   updatedBy,
			UpdatedAt:   now,
		})
	}

	product = p
	return
}
func (s *ServiceV1) Delete(ctx context.Context, id, by string) (err error) {
	ctx = s.uow.StartContext(ctx)
	defer s.uow.DeferErrorContext(ctx, err)
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("panic: %s", x)
			case error:
				err = fmt.Errorf("panic: %w", x)
			default:
				err = fmt.Errorf("panic: %v", x)
			}
		}
	}()

	err = s.repo.SoftDelete(ctx, id, by)
	if err != nil {
		return err
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.ProductDeletedEvent{
			ProductID: id,
			DeletedBy: by,
			DeletedAt: time.Now().UTC(),
		})
	}

	return nil
}
