package errors

import "net/http"

// HTTPStatusCode returns the appropriate HTTP status code for a domain error
func HTTPStatusCode(err error) int {
	var domainErr *DomainError
	if !As(err, &domainErr) {
		return http.StatusInternalServerError
	}

	switch domainErr.Code {
	case ErrNotFound.Code:
		return http.StatusNotFound
	case ErrAlreadyExists.Code:
		return http.StatusConflict
	case ErrConflict.Code:
		return http.StatusConflict
	case ErrUnauthorized.Code:
		return http.StatusUnauthorized
	case ErrForbidden.Code:
		return http.StatusForbidden
	case ErrInvalidToken.Code:
		return http.StatusUnauthorized
	case ErrInvalidCredentials.Code:
		return http.StatusUnauthorized
	case ErrValidation.Code:
		return http.StatusBadRequest
	case ErrInvalidInput.Code:
		return http.StatusBadRequest
	case ErrMissingField.Code:
		return http.StatusBadRequest
	case ErrInvalidFormat.Code:
		return http.StatusBadRequest
	case ErrBusinessRule.Code:
		return http.StatusUnprocessableEntity
	case ErrTimeout.Code:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// As is a convenience wrapper around errors.As
func As(err error, target any) bool {
	if err == nil {
		return false
	}

	switch v := target.(type) {
	case **DomainError:
		if de, ok := err.(*DomainError); ok {
			*v = de
			return true
		}
		// Check if wrapped
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			return As(wrapped.Unwrap(), target)
		}
	case **ValidationError:
		if ve, ok := err.(*ValidationError); ok {
			*v = ve
			return true
		}
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			return As(wrapped.Unwrap(), target)
		}
	}
	return false
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ToErrorResponse converts a domain error to an error response
func ToErrorResponse(err error) ErrorResponse {
	var domainErr *DomainError
	if As(err, &domainErr) {
		return ErrorResponse{
			Code:    domainErr.Code,
			Message: domainErr.Message,
		}
	}

	var validationErr *ValidationError
	if As(err, &validationErr) {
		return ErrorResponse{
			Code:    validationErr.Code,
			Message: validationErr.Message,
			Details: map[string]any{"fields": validationErr.Fields},
		}
	}

	return ErrorResponse{
		Code:    ErrInternal.Code,
		Message: "An unexpected error occurred",
	}
}
