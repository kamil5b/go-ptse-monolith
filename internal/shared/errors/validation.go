package errors

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error with field-level details
type ValidationError struct {
	*DomainError
	Fields map[string][]string // Field name -> list of validation messages
}

// NewValidationError creates a new validation error
func NewValidationError() *ValidationError {
	return &ValidationError{
		DomainError: ErrValidation,
		Fields:      make(map[string][]string),
	}
}

// AddFieldError adds a validation error for a specific field
func (e *ValidationError) AddFieldError(field, message string) *ValidationError {
	e.Fields[field] = append(e.Fields[field], message)
	return e
}

// HasErrors returns true if there are any validation errors
func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return e.DomainError.Error()
	}

	var msgs []string
	for field, errors := range e.Fields {
		for _, err := range errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", field, err))
		}
	}
	return fmt.Sprintf("%s: %s", e.Code, strings.Join(msgs, "; "))
}

// GetFieldErrors returns validation errors for a specific field
func (e *ValidationError) GetFieldErrors(field string) []string {
	return e.Fields[field]
}

// ToMap returns the validation errors as a map for JSON serialization
func (e *ValidationError) ToMap() map[string]any {
	return map[string]any{
		"code":    e.Code,
		"message": e.Message,
		"fields":  e.Fields,
	}
}
