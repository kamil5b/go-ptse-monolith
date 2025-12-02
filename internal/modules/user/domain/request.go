package domain

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required" validate:"required,min=1,max=255"`
	Email string `json:"email" binding:"required,email" validate:"required,email"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	ID    string `json:"id" binding:"required"`
	Name  string `json:"name" validate:"omitempty,min=1,max=255"`
	Email string `json:"email" binding:"omitempty,email" validate:"omitempty,email"`
}
