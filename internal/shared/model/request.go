package model

type PaginationRequest struct {
	Page  int `json:"page" binding:"required,min=1"`
	Limit int `json:"limit" binding:"required,min=5,max=100"`
}

func (p *PaginationRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}
