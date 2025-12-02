package domain

import "time"

// ProductResponse represents the product response payload
type ProductResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	CreatedBy   string     `json:"createdBy"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy   *string    `json:"updatedBy,omitempty"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
	DeletedBy   *string    `json:"deletedBy,omitempty"`
}

// ProductListResponse represents a list of products response
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
}

// ToResponse converts a Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		CreatedBy:   p.CreatedBy,
		UpdatedAt:   p.UpdatedAt,
		UpdatedBy:   p.UpdatedBy,
		DeletedAt:   p.DeletedAt,
		DeletedBy:   p.DeletedBy,
	}
}

// ToListResponse converts a slice of Products to ProductListResponse
func ToListResponse(products []Product) ProductListResponse {
	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = p.ToResponse()
	}
	return ProductListResponse{Products: responses}
}
