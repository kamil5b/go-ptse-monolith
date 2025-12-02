package v1

import (
	"context"
	"time"

	"go-modular-monolith/internal/modules/user/domain"
	"go-modular-monolith/internal/shared/events"
)

type ServiceV1 struct {
	repo     domain.Repository
	eventBus events.EventBus
}

func NewServiceV1(r domain.Repository, eb events.EventBus) *ServiceV1 {
	return &ServiceV1{repo: r, eventBus: eb}
}

func (s *ServiceV1) Create(ctx context.Context, req *domain.CreateUserRequest, createdBy string) (*domain.User, error) {
	ctx = s.repo.StartContext(ctx)
	var u domain.User
	u.Name = req.Name
	u.Email = req.Email
	u.CreatedAt = time.Now().UTC()
	u.CreatedBy = createdBy
	if err := s.repo.Create(ctx, &u); err != nil {
		s.repo.DeferErrorContext(ctx, err)
		return nil, err
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

	s.repo.DeferErrorContext(ctx, nil) // Commit transaction
	return &u, nil
}

func (s *ServiceV1) Get(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ServiceV1) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *ServiceV1) Update(ctx context.Context, req *domain.UpdateUserRequest, updatedBy string) (*domain.User, error) {
	ctx = s.repo.StartContext(ctx)
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
	if err := s.repo.Update(ctx, u); err != nil {
		s.repo.DeferErrorContext(ctx, err)
		return nil, err
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

	s.repo.DeferErrorContext(ctx, nil) // Commit transaction
	return u, nil
}

func (s *ServiceV1) Delete(ctx context.Context, id, by string) error {
	ctx = s.repo.StartContext(ctx)
	err := s.repo.SoftDelete(ctx, id, by)
	if err != nil {
		s.repo.DeferErrorContext(ctx, err)
		return err
	}

	// Publish event for inter-module communication
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, domain.UserDeletedEvent{
			UserID:    id,
			DeletedBy: by,
			DeletedAt: time.Now().UTC(),
		})
	}

	s.repo.DeferErrorContext(ctx, nil) // Commit transaction
	return nil
}
