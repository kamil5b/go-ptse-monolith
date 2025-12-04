package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDomainError(t *testing.T) {
	err := NewDomainError("TEST_CODE", "Test message")
	require.NotNil(t, err)
	assert.Equal(t, "TEST_CODE", err.Code)
	assert.Equal(t, "Test message", err.Message)
	assert.Nil(t, err.Err)
}

func TestDomainErrorError(t *testing.T) {
	err := NewDomainError("TEST_CODE", "Test message")
	assert.Equal(t, "TEST_CODE: Test message", err.Error())
}

func TestDomainErrorErrorWithUnderlying(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewDomainError("TEST_CODE", "Test message").WithError(underlying)
	assert.Contains(t, err.Error(), "TEST_CODE: Test message")
	assert.Contains(t, err.Error(), "underlying error")
}

func TestDomainErrorUnwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewDomainError("TEST_CODE", "Test message").WithError(underlying)
	assert.Equal(t, underlying, err.Unwrap())
}

func TestDomainErrorWithMessage(t *testing.T) {
	err := NewDomainError("TEST_CODE", "Original message")
	newErr := err.WithMessage("New message")

	assert.Equal(t, "TEST_CODE", newErr.Code)
	assert.Equal(t, "New message", newErr.Message)
	assert.Equal(t, "Original message", err.Message)
}

func TestDomainErrorWithError(t *testing.T) {
	underlying := errors.New("underlying")
	err := NewDomainError("TEST_CODE", "Test message")
	newErr := err.WithError(underlying)

	assert.Equal(t, "TEST_CODE", newErr.Code)
	assert.Equal(t, "Test message", newErr.Message)
	assert.Equal(t, underlying, newErr.Err)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *DomainError
		code string
	}{
		{"NotFound", ErrNotFound, "NOT_FOUND"},
		{"AlreadyExists", ErrAlreadyExists, "ALREADY_EXISTS"},
		{"Conflict", ErrConflict, "CONFLICT"},
		{"Unauthorized", ErrUnauthorized, "UNAUTHORIZED"},
		{"Forbidden", ErrForbidden, "FORBIDDEN"},
		{"InvalidToken", ErrInvalidToken, "INVALID_TOKEN"},
		{"InvalidCredentials", ErrInvalidCredentials, "INVALID_CREDENTIALS"},
		{"Validation", ErrValidation, "VALIDATION_ERROR"},
		{"InvalidInput", ErrInvalidInput, "INVALID_INPUT"},
		{"MissingField", ErrMissingField, "MISSING_FIELD"},
		{"InvalidFormat", ErrInvalidFormat, "INVALID_FORMAT"},
		{"BusinessRule", ErrBusinessRule, "BUSINESS_RULE_VIOLATION"},
		{"OperationFailed", ErrOperationFailed, "OPERATION_FAILED"},
		{"Internal", ErrInternal, "INTERNAL_ERROR"},
		{"DatabaseError", ErrDatabaseError, "DATABASE_ERROR"},
		{"ExternalService", ErrExternalService, "EXTERNAL_SERVICE_ERROR"},
		{"Timeout", ErrTimeout, "TIMEOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

func TestIsFunction(t *testing.T) {
	err := ErrNotFound.WithError(errors.New("test"))
	assert.True(t, Is(err, ErrNotFound))
	assert.False(t, Is(err, ErrForbidden))
}

func TestCodeFunction(t *testing.T) {
	err := ErrNotFound
	assert.Equal(t, "NOT_FOUND", Code(err))

	wrappedErr := ErrNotFound.WithError(errors.New("test"))
	assert.Equal(t, "NOT_FOUND", Code(wrappedErr))

	unknownErr := errors.New("unknown")
	assert.Equal(t, "UNKNOWN_ERROR", Code(unknownErr))
}

func TestDomainErrorChaining(t *testing.T) {
	baseErr := errors.New("base error")
	domainErr := ErrDatabaseError.
		WithError(baseErr).
		WithMessage("Failed to query database")

	assert.Equal(t, "DATABASE_ERROR", domainErr.Code)
	assert.Equal(t, "Failed to query database", domainErr.Message)
	assert.Equal(t, baseErr, domainErr.Err)
}

func TestErrorsAsSupport(t *testing.T) {
	underlying := errors.New("test error")
	err := NewDomainError("TEST", "message").WithError(underlying)

	var domainErr *DomainError
	assert.True(t, errors.As(err, &domainErr))
	assert.Equal(t, "TEST", domainErr.Code)
}
