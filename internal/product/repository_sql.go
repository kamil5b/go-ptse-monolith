package product

import (
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

func (r *SQLRepository) Create(p *Product) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.CreatedAt = time.Now().UTC()
	_, err := r.db.NamedExec(`INSERT INTO products (id,name,description,created_at,created_by) VALUES (:id,:name,:description,:created_at,:created_by)`, p)
	return err
}

func (r *SQLRepository) GetByID(id string) (*Product, error) {
	var p Product
	if err := r.db.Get(&p, `SELECT id,name,description,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by FROM products WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SQLRepository) List() ([]Product, error) {
	var lst []Product
	if err := r.db.Select(&lst, `SELECT id,name,description,created_at,created_by,updated_at,updated_by FROM products WHERE deleted_at IS NULL ORDER BY created_at DESC`); err != nil {
		return nil, err
	}
	return lst, nil
}

func (r *SQLRepository) Update(p *Product) error {
	now := time.Now().UTC()
	p.UpdatedAt = &now
	_, err := r.db.NamedExec(`UPDATE products SET name=:name, description=:description, updated_at=:updated_at, updated_by=:updated_by WHERE id=:id`, p)
	return err
}

func (r *SQLRepository) SoftDelete(id, deletedBy string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`UPDATE products SET deleted_at=$1, deleted_by=$2 WHERE id=$3`, now, deletedBy, id)
	return err
}
