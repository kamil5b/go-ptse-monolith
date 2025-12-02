package noop

import (
	"context"
	"errors"

	"go-modular-monolith/internal/modules/product/domain"
)

type UnimplementedRepository struct{}

func NewUnimplementedRepository() *UnimplementedRepository {
	return &UnimplementedRepository{}
}

func (s *UnimplementedRepository) Create(_ context.Context, _ *domain.Product) error {
	return errors.New("not implemented")
}
func (s *UnimplementedRepository) GetByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedRepository) List(_ context.Context) ([]domain.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedRepository) Update(_ context.Context, _ *domain.Product) error {
	return errors.New("not implemented")
}
func (s *UnimplementedRepository) SoftDelete(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}
