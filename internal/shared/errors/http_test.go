package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{"NotFound", ErrNotFound, http.StatusNotFound},
		{"AlreadyExists", ErrAlreadyExists, http.StatusConflict},
		{"Conflict", ErrConflict, http.StatusConflict},
		{"Unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, http.StatusForbidden},
		{"InvalidToken", ErrInvalidToken, http.StatusUnauthorized},
		{"InvalidCredentials", ErrInvalidCredentials, http.StatusUnauthorized},
		{"Validation", ErrValidation, http.StatusBadRequest},
		{"InvalidInput", ErrInvalidInput, http.StatusBadRequest},
		{"MissingField", ErrMissingField, http.StatusBadRequest},
		{"InvalidFormat", ErrInvalidFormat, http.StatusBadRequest},
		{"BusinessRule", ErrBusinessRule, http.StatusUnprocessableEntity},
		{"Timeout", ErrTimeout, http.StatusGatewayTimeout},
		{"Internal", ErrInternal, http.StatusInternalServerError},
		{"DatabaseError", ErrDatabaseError, http.StatusInternalServerError},
		{"UnknownError", errors.New("unknown"), http.StatusInternalServerError},
		{"NilError", nil, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.statusCode, HTTPStatusCode(tt.err))
		})
	}
}

func TestHTTPStatusCodeWithWrappedError(t *testing.T) {
	wrapped := ErrNotFound.WithError(errors.New("details"))
	assert.Equal(t, http.StatusNotFound, HTTPStatusCode(wrapped))
}

func TestAsWithDomainError(t *testing.T) {
	err := NewDomainError("TEST", "message")
	var domainErr *DomainError
	result := As(err, &domainErr)
	assert.True(t, result)
	assert.Equal(t, "TEST", domainErr.Code)
}

func TestAsWithWrappedDomainError(t *testing.T) {
	underlying := ErrNotFound.WithError(errors.New("test"))
	var domainErr *DomainError
	result := As(underlying, &domainErr)
	assert.True(t, result)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
}

func TestAsWithNonDomainError(t *testing.T) {
	err := errors.New("generic error")
	var domainErr *DomainError
	result := As(err, &domainErr)
	assert.False(t, result)
}

func TestAsWithNilError(t *testing.T) {
	var domainErr *DomainError
	result := As(nil, &domainErr)
	assert.False(t, result)
}

func TestToErrorResponseWithDomainError(t *testing.T) {
	err := NewDomainError("TEST_CODE", "Test message")
	response := ToErrorResponse(err)

	assert.Equal(t, "TEST_CODE", response.Code)
	assert.Equal(t, "Test message", response.Message)
	assert.Nil(t, response.Details)
}

func TestToErrorResponseWithWrappedDomainError(t *testing.T) {
	err := ErrValidation.WithError(errors.New("validation failed"))
	response := ToErrorResponse(err)

	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.NotEmpty(t, response.Message)
}

func TestToErrorResponseWithGenericError(t *testing.T) {
	err := errors.New("generic error")
	response := ToErrorResponse(err)

	assert.Equal(t, "INTERNAL_ERROR", response.Code)
	assert.Equal(t, "An unexpected error occurred", response.Message)
}

func TestToErrorResponseWithNilError(t *testing.T) {
	response := ToErrorResponse(nil)

	assert.Equal(t, "INTERNAL_ERROR", response.Code)
	assert.Equal(t, "An unexpected error occurred", response.Message)
}

func TestErrorResponseStructure(t *testing.T) {
	response := ErrorResponse{
		Code:    "TEST",
		Message: "Test message",
		Details: map[string]any{"key": "value"},
	}

	assert.Equal(t, "TEST", response.Code)
	assert.Equal(t, "Test message", response.Message)
	assert.NotNil(t, response.Details)
	assert.Equal(t, "value", response.Details["key"])
}
