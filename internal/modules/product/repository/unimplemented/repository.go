package repository

import (
	"errors"
	"go-modular-monolith/internal/domain/product"
)

type UnimplementedRepository struct{}

func NewUnimplementedRepository() *UnimplementedRepository {
	return &UnimplementedRepository{}
}

func (s *UnimplementedRepository) Create(p *product.Product) error {
	return errors.New("not implemented")
}
func (s *UnimplementedRepository) GetByID(id string) (*product.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedRepository) List() ([]product.Product, error) {
	return nil, errors.New("not implemented")
}
func (s *UnimplementedRepository) Update(p *product.Product) error {
	return errors.New("not implemented")
}
func (s *UnimplementedRepository) SoftDelete(id, deletedBy string) error {
	return errors.New("not implemented")
}
