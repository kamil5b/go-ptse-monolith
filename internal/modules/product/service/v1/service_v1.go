package v1

import (
	"context"
	"fmt"
	"go-modular-monolith/internal/domain/product"

	"github.com/sirupsen/logrus"
)

type ServiceV1 struct {
	repo product.ProductRepository
}

func NewServiceV1(r product.ProductRepository) *ServiceV1 { return &ServiceV1{repo: r} }

func (s *ServiceV1) Create(ctx context.Context, p *product.Product) (err error) {
	ctx = s.repo.StartContext(ctx)
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("panic recovered in product service v1 Create: %v", panicErr)
			logrus.Error(err)
		}
		s.repo.DeferErrorContext(ctx, err)
	}()
	err = s.repo.Create(ctx, p)
	if err != nil {
		return err
	}
	return nil
}
func (s *ServiceV1) Get(ctx context.Context, id string) (*product.Product, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *ServiceV1) List(ctx context.Context) ([]product.Product, error) {
	return s.repo.List(ctx)
}
func (s *ServiceV1) Update(ctx context.Context, p *product.Product) error {
	return s.repo.Update(ctx, p)
}
func (s *ServiceV1) Delete(ctx context.Context, id, by string) error {
	return s.repo.SoftDelete(ctx, id, by)
}
