package domain

import "time"

// ProductCreatedEvent is published when a new product is created
type ProductCreatedEvent struct {
	ProductID   string    `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e ProductCreatedEvent) EventName() string { return "product.created" }
func (e ProductCreatedEvent) Payload() any      { return e }

// ProductUpdatedEvent is published when a product is updated
type ProductUpdatedEvent struct {
	ProductID   string    `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (e ProductUpdatedEvent) EventName() string { return "product.updated" }
func (e ProductUpdatedEvent) Payload() any      { return e }

// ProductDeletedEvent is published when a product is soft-deleted
type ProductDeletedEvent struct {
	ProductID string    `json:"product_id"`
	DeletedBy string    `json:"deleted_by"`
	DeletedAt time.Time `json:"deleted_at"`
}

func (e ProductDeletedEvent) EventName() string { return "product.deleted" }
func (e ProductDeletedEvent) Payload() any      { return e }
