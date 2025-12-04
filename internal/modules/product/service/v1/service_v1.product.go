package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/cache"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/events"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/uow"
)

const (
	productCacheKeyPrefix = "product:"
	productCacheTTL       = 15 * time.Minute
)

type ServiceV1 struct {
	repo     domain.Repository
	uow      uow.UnitOfWork
	eventBus events.EventBus
	cache    cache.Cache
}

func NewServiceV1(r domain.Repository, u uow.UnitOfWork, eb events.EventBus, c cache.Cache) *ServiceV1 {
	return &ServiceV1{repo: r, uow: u, eventBus: eb, cache: c}
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
	// Try to get from cache first
	if s.cache != nil {
		cacheKey := productCacheKeyPrefix + id
		if cached, err := s.cache.GetBytes(ctx, cacheKey); err == nil {
			var product domain.Product
			if err := json.Unmarshal(cached, &product); err == nil {
				return &product, nil
			}
		}
	}

	// Cache miss - fetch from repository
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if s.cache != nil && product != nil {
		cacheKey := productCacheKeyPrefix + id
		if data, err := json.Marshal(product); err == nil {
			_ = s.cache.Set(ctx, cacheKey, data, productCacheTTL)
		}
	}

	return product, nil
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

	// Invalidate cache after update
	if s.cache != nil {
		cacheKey := productCacheKeyPrefix + p.ID
		_ = s.cache.Delete(ctx, cacheKey)
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

	// Invalidate cache after delete
	if s.cache != nil {
		cacheKey := productCacheKeyPrefix + id
		_ = s.cache.Delete(ctx, cacheKey)
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
