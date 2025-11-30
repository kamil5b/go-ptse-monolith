package unimplemented

import (
	"context"
	"errors"
	"go-modular-monolith/internal/domain/product"
)

type UnimplementedService struct{}

func NewUnimplementedService() *UnimplementedService {
	return &UnimplementedService{}
}

func (s *UnimplementedService) Create(_ context.Context, _ *product.Product) error {
	return errors.New("not implemented")
}
func (s *UnimplementedService) Get(_ context.Context, _ string) (*product.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedService) List(_ context.Context) ([]product.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedService) Update(_ context.Context, _ *product.Product) error {
	return errors.New("not implemented")
}
func (s *UnimplementedService) Delete(_ context.Context, _ string, _ string) error {
	return errors.New("not implemented")
}
