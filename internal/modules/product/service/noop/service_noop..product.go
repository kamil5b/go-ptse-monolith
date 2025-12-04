package noop

import (
	"context"
	"errors"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
)

type UnimplementedService struct{}

func NewUnimplementedService() *UnimplementedService {
	return &UnimplementedService{}
}

func (s *UnimplementedService) Create(_ context.Context, _ *domain.CreateProductRequest, _ string) (*domain.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedService) Get(_ context.Context, _ string) (*domain.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedService) List(_ context.Context) ([]domain.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedService) Update(_ context.Context, _ *domain.UpdateProductRequest, _ string) (*domain.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedService) Delete(_ context.Context, _ string, _ string) error {
	return errors.New("not implemented")
}
