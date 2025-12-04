package storage

import "fmt"

// StorageErrorType defines the type of storage error
type StorageErrorType string

const (
	ErrTypeNotFound          StorageErrorType = "STORAGE_NOT_FOUND"
	ErrTypeSizeLimitExceeded StorageErrorType = "STORAGE_SIZE_LIMIT_EXCEEDED"
	ErrTypePermissionDenied  StorageErrorType = "STORAGE_PERMISSION_DENIED"
	ErrTypeInvalidPath       StorageErrorType = "STORAGE_INVALID_PATH"
	ErrTypeServiceError      StorageErrorType = "STORAGE_SERVICE_ERROR"
	ErrTypeUnknown           StorageErrorType = "STORAGE_UNKNOWN"
)

// StorageError represents a storage-related error
type StorageError struct {
	Type     StorageErrorType
	Message  string
	Err      error
	HTTPCode int
}

// Error implements the error interface
func (e *StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *StorageError) Unwrap() error {
	return e.Err
}

// NotFound creates a NOT_FOUND error
func NotFound(path string) *StorageError {
	return &StorageError{
		Type:     ErrTypeNotFound,
		Message:  fmt.Sprintf("object not found: %s", path),
		HTTPCode: 404,
	}
}

// SizeLimitExceeded creates a SIZE_LIMIT_EXCEEDED error
func SizeLimitExceeded(maxSize int64) *StorageError {
	return &StorageError{
		Type:     ErrTypeSizeLimitExceeded,
		Message:  fmt.Sprintf("file size exceeds limit of %d bytes", maxSize),
		HTTPCode: 413,
	}
}

// PermissionDenied creates a PERMISSION_DENIED error
func PermissionDenied(path string) *StorageError {
	return &StorageError{
		Type:     ErrTypePermissionDenied,
		Message:  fmt.Sprintf("permission denied: %s", path),
		HTTPCode: 403,
	}
}

// InvalidPath creates an INVALID_PATH error
func InvalidPath(path string) *StorageError {
	return &StorageError{
		Type:     ErrTypeInvalidPath,
		Message:  fmt.Sprintf("invalid path: %s", path),
		HTTPCode: 400,
	}
}

// ServiceError creates a SERVICE_ERROR
func ServiceError(msg string, err error) *StorageError {
	return &StorageError{
		Type:     ErrTypeServiceError,
		Message:  msg,
		Err:      err,
		HTTPCode: 500,
	}
}
