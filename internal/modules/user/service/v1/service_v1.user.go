package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/cache"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/email"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/events"
)

const (
	userCacheKeyPrefix = "user:"
	userCacheTTL       = 15 * time.Minute
)

type ServiceV1 struct {
	repo         domain.Repository
	eventBus     events.EventBus
	emailService email.EmailService
	cache        cache.Cache
}

func NewServiceV1(r domain.Repository, eb events.EventBus, es email.EmailService, c cache.Cache) *ServiceV1 {
	return &ServiceV1{repo: r, eventBus: eb, emailService: es, cache: c}
}

func (s *ServiceV1) Create(ctx context.Context, req *domain.CreateUserRequest, createdBy string) (user *domain.User, err error) {
	ctx = s.repo.StartContext(ctx)
	defer s.repo.DeferErrorContext(ctx, err)
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

	var u domain.User
	u.Name = req.Name
	u.Email = req.Email
	u.CreatedAt = time.Now().UTC()
	u.CreatedBy = createdBy
	if err = s.repo.Create(ctx, &u); err != nil {
		return nil, err
	}

	// Send welcome email asynchronously (via worker or background task)
	if s.emailService != nil {
		_ = s.emailService.Send(ctx, &email.Email{
			To:       []string{u.Email},
			Subject:  fmt.Sprintf("Welcome %s", u.Name),
			TextBody: fmt.Sprintf("Welcome to our platform, %s!\n\nYour account has been created successfully.", u.Name),
			HTMLBody: fmt.Sprintf("<h1>Welcome %s</h1><p>Your account has been created successfully.</p>", u.Name),
		})
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.UserCreatedEvent{
			UserID:    u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedBy: createdBy,
			CreatedAt: u.CreatedAt,
		})
	}

	user = &u
	return
}

func (s *ServiceV1) Get(ctx context.Context, id string) (*domain.User, error) {
	// Try to get from cache first
	if s.cache != nil {
		cacheKey := userCacheKeyPrefix + id
		if cached, err := s.cache.GetBytes(ctx, cacheKey); err == nil {
			var user domain.User
			if err := json.Unmarshal(cached, &user); err == nil {
				return &user, nil
			}
		}
	}

	// Cache miss - fetch from repository
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if s.cache != nil && user != nil {
		cacheKey := userCacheKeyPrefix + id
		if data, err := json.Marshal(user); err == nil {
			_ = s.cache.Set(ctx, cacheKey, data, userCacheTTL)
		}
	}

	return user, nil
}

func (s *ServiceV1) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *ServiceV1) Update(ctx context.Context, req *domain.UpdateUserRequest, updatedBy string) (user *domain.User, err error) {
	ctx = s.repo.StartContext(ctx)
	defer s.repo.DeferErrorContext(ctx, err)
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

	u, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	now := time.Now().UTC()
	u.UpdatedAt = &now
	u.UpdatedBy = &updatedBy
	if err = s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	// Invalidate cache after update
	if s.cache != nil {
		cacheKey := userCacheKeyPrefix + u.ID
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.UserUpdatedEvent{
			UserID:    u.ID,
			Name:      u.Name,
			Email:     u.Email,
			UpdatedBy: updatedBy,
			UpdatedAt: now,
		})
	}

	user = u
	return
}

func (s *ServiceV1) Delete(ctx context.Context, id, by string) (err error) {
	ctx = s.repo.StartContext(ctx)
	defer s.repo.DeferErrorContext(ctx, err)
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
		cacheKey := userCacheKeyPrefix + id
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.UserDeletedEvent{
			UserID:    id,
			DeletedBy: by,
			DeletedAt: time.Now().UTC(),
		})
	}

	return nil
}
