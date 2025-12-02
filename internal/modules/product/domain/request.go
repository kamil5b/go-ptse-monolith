package domain

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Name        string `json:"name" binding:"required" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" validate:"omitempty,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}
