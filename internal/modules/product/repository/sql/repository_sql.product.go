package sql

import (
	"context"
	"go-modular-monolith/internal/modules/product/domain"
	sharedCtx "go-modular-monolith/internal/shared/context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) getTxFromContext(ctx context.Context) *sqlx.Tx {
	return sharedCtx.GetObjectFromContext[sqlx.Tx](ctx, sharedCtx.PostgresTxKey)
}

func (r *SQLRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (id,name,description,created_at,created_by) VALUES (:id,:name,:description,:created_at,:created_by)`
	tx := r.getTxFromContext(ctx)
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.CreatedAt = time.Now().UTC()
	if tx != nil {
		_, err := tx.NamedExec(query, p)
		return err
	}
	_, err := r.db.NamedExec(query, p)
	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	tx := r.getTxFromContext(ctx)
	query := `SELECT id,name,description,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by FROM products WHERE id=$1`
	if tx != nil {
		if err := tx.Get(&p, query, id); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&p, query, id); err != nil {
			return nil, err
		}
	}
	return &p, nil
}

func (r *SQLRepository) List(ctx context.Context) ([]domain.Product, error) {
	var lst []domain.Product
	tx := r.getTxFromContext(ctx)
	query := `SELECT id,name,description,created_at,created_by,updated_at,updated_by FROM products WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if tx != nil {
		if err := tx.Select(&lst, query); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Select(&lst, query); err != nil {
			return nil, err
		}
	}
	return lst, nil
}

func (r *SQLRepository) Update(ctx context.Context, p *domain.Product) error {
	now := time.Now().UTC()
	p.UpdatedAt = &now
	tx := r.getTxFromContext(ctx)
	query := `UPDATE products SET name=:name, description=:description, updated_at=:updated_at, updated_by=:updated_by WHERE id=:id`
	if tx != nil {
		_, err := tx.NamedExec(query, p)
		return err
	}
	_, err := r.db.NamedExec(query, p)
	return err
}

func (r *SQLRepository) SoftDelete(ctx context.Context, id, deletedBy string) error {
	now := time.Now().UTC()
	tx := r.getTxFromContext(ctx)
	query := `UPDATE products SET deleted_at=$1, deleted_by=$2 WHERE id=$3`
	if tx != nil {
		_, err := tx.Exec(query, now, deletedBy, id)
		return err
	}
	_, err := r.db.Exec(query, now, deletedBy, id)
	return err
}
