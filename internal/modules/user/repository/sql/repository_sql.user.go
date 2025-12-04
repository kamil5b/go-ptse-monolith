package sql

import (
	"context"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/user/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const userDriverName = "UserPostgreSQL"

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) StartContext(ctx context.Context) context.Context {
	tx := r.db.MustBeginTx(ctx, nil)
	return context.WithValue(ctx, userDriverName, tx)
}

func (r *SQLRepository) DeferErrorContext(ctx context.Context, err error) {
	tx := r.getTxFromContext(ctx)
	if tx != nil {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}
}

func (r *SQLRepository) getTxFromContext(ctx context.Context) *sqlx.Tx {
	txVal := ctx.Value(userDriverName)
	tx, ok := txVal.(*sqlx.Tx)
	if !ok {
		return nil
	}
	return tx
}

func (r *SQLRepository) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id,name,email,created_at,created_by) VALUES (:id,:name,:email,:created_at,:created_by)`
	tx := r.getTxFromContext(ctx)
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	u.CreatedAt = time.Now().UTC()
	if tx != nil {
		_, err := tx.NamedExec(query, u)
		return err
	}
	_, err := r.db.NamedExec(query, u)
	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	tx := r.getTxFromContext(ctx)
	query := `SELECT id,name,email,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by FROM users WHERE id=$1`
	if tx != nil {
		if err := tx.Get(&u, query, id); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&u, query, id); err != nil {
			return nil, err
		}
	}
	return &u, nil
}

func (r *SQLRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	tx := r.getTxFromContext(ctx)
	query := `SELECT id,name,email,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by FROM users WHERE email=$1 AND deleted_at IS NULL`
	if tx != nil {
		if err := tx.Get(&u, query, email); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&u, query, email); err != nil {
			return nil, err
		}
	}
	return &u, nil
}

func (r *SQLRepository) List(ctx context.Context) ([]domain.User, error) {
	var lst []domain.User
	tx := r.getTxFromContext(ctx)
	query := `SELECT id,name,email,created_at,created_by,updated_at,updated_by FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC`
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

func (r *SQLRepository) Update(ctx context.Context, u *domain.User) error {
	now := time.Now().UTC()
	u.UpdatedAt = &now
	tx := r.getTxFromContext(ctx)
	query := `UPDATE users SET name=:name, email=:email, updated_at=:updated_at, updated_by=:updated_by WHERE id=:id`
	if tx != nil {
		_, err := tx.NamedExec(query, u)
		return err
	}
	_, err := r.db.NamedExec(query, u)
	return err
}

func (r *SQLRepository) SoftDelete(ctx context.Context, id, deletedBy string) error {
	now := time.Now().UTC()
	tx := r.getTxFromContext(ctx)
	query := `UPDATE users SET deleted_at=$1, deleted_by=$2 WHERE id=$3`
	if tx != nil {
		_, err := tx.Exec(query, now, deletedBy, id)
		return err
	}
	_, err := r.db.Exec(query, now, deletedBy, id)
	return err
}
