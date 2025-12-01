package uow

import "context"

type UnitOfWork interface {
	StartContext(ctx context.Context) context.Context
	DeferErrorContext(ctx context.Context, err error) error
}
