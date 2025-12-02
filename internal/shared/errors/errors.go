package errors

import (
	"errors"
	"fmt"
)

// DomainError represents a structured error for the domain layer
type DomainError struct {
	Code    string // Machine-readable error code
	Message string // Human-readable message
	Err     error  // Underlying error (optional)
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *DomainError) Unwrap() error {
	return e.Err
}

// WithError wraps an underlying error
func (e *DomainError) WithError(err error) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// WithMessage creates a new error with a custom message
func (e *DomainError) WithMessage(msg string) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: msg,
		Err:     e.Err,
	}
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// Pre-defined domain errors
var (
	// Resource errors
	ErrNotFound      = NewDomainError("NOT_FOUND", "resource not found")
	ErrAlreadyExists = NewDomainError("ALREADY_EXISTS", "resource already exists")
	ErrConflict      = NewDomainError("CONFLICT", "resource conflict")

	// Authentication errors
	ErrUnauthorized       = NewDomainError("UNAUTHORIZED", "unauthorized access")
	ErrForbidden          = NewDomainError("FORBIDDEN", "access forbidden")
	ErrInvalidToken       = NewDomainError("INVALID_TOKEN", "invalid or expired token")
	ErrInvalidCredentials = NewDomainError("INVALID_CREDENTIALS", "invalid credentials")

	// Validation errors
	ErrValidation    = NewDomainError("VALIDATION_ERROR", "validation failed")
	ErrInvalidInput  = NewDomainError("INVALID_INPUT", "invalid input provided")
	ErrMissingField  = NewDomainError("MISSING_FIELD", "required field is missing")
	ErrInvalidFormat = NewDomainError("INVALID_FORMAT", "invalid format")

	// Business logic errors
	ErrBusinessRule    = NewDomainError("BUSINESS_RULE_VIOLATION", "business rule violation")
	ErrOperationFailed = NewDomainError("OPERATION_FAILED", "operation failed")

	// Infrastructure errors
	ErrInternal        = NewDomainError("INTERNAL_ERROR", "internal server error")
	ErrDatabaseError   = NewDomainError("DATABASE_ERROR", "database operation failed")
	ErrExternalService = NewDomainError("EXTERNAL_SERVICE_ERROR", "external service error")
	ErrTimeout         = NewDomainError("TIMEOUT", "operation timed out")
)

// Is checks if the target error matches the domain error code
func Is(err error, target *DomainError) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == target.Code
	}
	return false
}

// Code extracts the error code from an error
func Code(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return "UNKNOWN_ERROR"
}
