package product

import "time"

type Product struct {
	ID          string     `db:"id" json:"id" bson:"id"`
	Name        string     `db:"name" json:"name" bson:"name"`
	Description string     `db:"description" json:"description" bson:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at" bson:"created_at"`
	CreatedBy   string     `db:"created_by" json:"created_by" bson:"created_by"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	UpdatedBy   *string    `db:"updated_by" json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy   *string    `db:"deleted_by" json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
}
