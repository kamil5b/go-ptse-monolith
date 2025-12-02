package model

type CommonResponse struct {
	RequestID string `json:"requestId"`
}

type PaginationMetadata struct {
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
}

type PaginatedResponse[T any] struct {
	CommonResponse
	Metadata PaginationMetadata `json:"metadata"`
	Data     []T                `json:"data"`
}

func NewPaginatedResponse[T any](requestID string, data []T, totalItems int, meta PaginationMetadata) *PaginatedResponse[T] {
	totalPages := (totalItems + meta.Limit - 1) / meta.Limit

	return &PaginatedResponse[T]{
		CommonResponse: CommonResponse{
			RequestID: requestID,
		},
		Metadata: PaginationMetadata{
			TotalItems: totalItems,
			TotalPages: totalPages,
			Page:       meta.Page,
			Limit:      meta.Limit,
		},
		Data: data,
	}
}
