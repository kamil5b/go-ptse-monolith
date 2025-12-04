package unitofwork

import (
	"context"

	sharedCtx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"

	"github.com/jmoiron/sqlx"
)

type SQLUnitOfWork struct {
	db *sqlx.DB
}

func NewSQLUnitOfWork(db *sqlx.DB) *SQLUnitOfWork {
	return &SQLUnitOfWork{db: db}
}

func (r *SQLUnitOfWork) StartContext(ctx context.Context) context.Context {
	tx := r.db.MustBeginTx(ctx, nil)
	return context.WithValue(ctx, sharedCtx.PostgresTxKey, &tx)
}

func (r *SQLUnitOfWork) DeferErrorContext(ctx context.Context, err error) error {
	tx := sharedCtx.GetObjectFromContext[sqlx.Tx](ctx, sharedCtx.PostgresTxKey)
	if tx != nil {
		if err != nil {
			return tx.Rollback()
		} else {
			return tx.Commit()
		}
	}
	return nil
}
