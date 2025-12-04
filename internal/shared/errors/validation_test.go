package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError()
	require.NotNil(t, err)
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.NotNil(t, err.Fields)
	assert.Empty(t, err.Fields)
}

func TestAddFieldError(t *testing.T) {
	err := NewValidationError()
	result := err.AddFieldError("email", "must be a valid email")

	assert.Equal(t, err, result) // Should return self for chaining
	assert.Equal(t, 1, len(err.Fields))
	assert.Equal(t, []string{"must be a valid email"}, err.Fields["email"])
}

func TestAddMultipleFieldErrors(t *testing.T) {
	err := NewValidationError()
	err.AddFieldError("email", "must be a valid email")
	err.AddFieldError("email", "must not be empty")
	err.AddFieldError("password", "must be at least 8 characters")

	assert.Equal(t, 2, len(err.Fields))
	assert.Equal(t, 2, len(err.Fields["email"]))
	assert.Equal(t, 1, len(err.Fields["password"]))
	assert.Contains(t, err.Fields["email"], "must be a valid email")
	assert.Contains(t, err.Fields["email"], "must not be empty")
}

func TestHasErrors(t *testing.T) {
	err := NewValidationError()
	assert.False(t, err.HasErrors())

	err.AddFieldError("field", "error")
	assert.True(t, err.HasErrors())
}

func TestValidationErrorString(t *testing.T) {
	err := NewValidationError()
	err.AddFieldError("email", "must be valid")
	err.AddFieldError("password", "must be secure")

	errStr := err.Error()
	assert.Contains(t, errStr, "VALIDATION_ERROR")
	assert.Contains(t, errStr, "email: must be valid")
	assert.Contains(t, errStr, "password: must be secure")
}

func TestValidationErrorStringEmpty(t *testing.T) {
	err := NewValidationError()
	errStr := err.Error()
	assert.Equal(t, "VALIDATION_ERROR: validation failed", errStr)
}

func TestGetFieldErrors(t *testing.T) {
	err := NewValidationError()
	err.AddFieldError("username", "already exists")
	err.AddFieldError("username", "must be alphanumeric")

	errors := err.GetFieldErrors("username")
	assert.Equal(t, 2, len(errors))
	assert.Equal(t, "already exists", errors[0])
	assert.Equal(t, "must be alphanumeric", errors[1])
}

func TestGetFieldErrorsNotFound(t *testing.T) {
	err := NewValidationError()
	errors := err.GetFieldErrors("nonexistent")
	assert.Nil(t, errors)
}

func TestToMap(t *testing.T) {
	err := NewValidationError()
	err.AddFieldError("email", "invalid")
	err.AddFieldError("age", "must be positive")

	m := err.ToMap()
	assert.Equal(t, "VALIDATION_ERROR", m["code"])
	assert.NotNil(t, m["message"])
	assert.NotNil(t, m["fields"])

	fields := m["fields"].(map[string][]string)
	assert.Equal(t, []string{"invalid"}, fields["email"])
	assert.Equal(t, []string{"must be positive"}, fields["age"])
}

func TestValidationErrorChaining(t *testing.T) {
	err := NewValidationError().
		AddFieldError("email", "must be valid").
		AddFieldError("password", "too weak").
		AddFieldError("password", "must contain numbers")

	assert.Equal(t, 2, len(err.Fields))
	assert.Equal(t, 2, len(err.Fields["password"]))
}

func TestValidationErrorAsInterface(t *testing.T) {
	err := NewValidationError()
	err.AddFieldError("field", "error")

	// Should implement error interface
	var e error = err
	assert.NotNil(t, e.Error())
}

func TestValidationErrorWithDomainErrorFields(t *testing.T) {
	err := NewValidationError()
	assert.Equal(t, "VALIDATION_ERROR", err.DomainError.Code)
	assert.NotEmpty(t, err.DomainError.Message)
}
