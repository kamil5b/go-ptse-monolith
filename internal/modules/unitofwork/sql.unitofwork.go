package unitofwork

import (
	"context"
	"go-modular-monolith/pkg/constant"
	"go-modular-monolith/pkg/util"

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
	return context.WithValue(ctx, constant.ContextKeyPostgresTx, &tx)
}

func (r *SQLUnitOfWork) DeferErrorContext(ctx context.Context, err error) error {
	tx := util.GetObjectFromContext[sqlx.Tx](ctx, constant.ContextKeyPostgresTx)
	if tx != nil {
		if err != nil {
			return tx.Rollback()
		} else {
			return tx.Commit()
		}
	}
	return nil
}
