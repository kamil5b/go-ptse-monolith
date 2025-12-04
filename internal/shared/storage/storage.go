package storage

import (
	"context"
	"io"
	"time"
)

// StorageObject represents metadata about a stored object
type StorageObject struct {
	Name         string            // Object name/path
	Size         int64             // File size in bytes
	ContentType  string            // MIME type
	ETag         string            // Entity tag (MD5 or hash)
	LastModified time.Time         // Last modification time
	Metadata     map[string]string // Custom metadata
	PresignedURL string            // Presigned URL (if applicable)
}

// UploadOptions configures upload behavior
type UploadOptions struct {
	ContentType          string            // MIME type
	Metadata             map[string]string // Custom metadata
	CacheControl         string            // Cache-Control header
	ContentEncoding      string            // Content-Encoding header
	ACL                  string            // Access control (private/public-read)
	ServerSideEncryption bool              // Enable encryption
}

// StorageService provides unified file storage operations
type StorageService interface {
	// Upload stores a file and returns metadata
	Upload(ctx context.Context, path string, reader io.Reader, opts *UploadOptions) (*StorageObject, error)

	// UploadBytes stores bytes and returns metadata
	UploadBytes(ctx context.Context, path string, data []byte, opts *UploadOptions) (*StorageObject, error)

	// Download retrieves a file
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// GetBytes retrieves file contents as bytes
	GetBytes(ctx context.Context, path string) ([]byte, error)

	// GetObject retrieves object metadata
	GetObject(ctx context.Context, path string) (*StorageObject, error)

	// Delete removes a file
	Delete(ctx context.Context, path string) error

	// DeletePrefix removes all objects with given prefix
	DeletePrefix(ctx context.Context, prefix string) error

	// Exists checks if object exists
	Exists(ctx context.Context, path string) (bool, error)

	// ListObjects lists objects in a path prefix
	ListObjects(ctx context.Context, prefix string, recursive bool) ([]*StorageObject, error)

	// GetPresignedURL generates a temporary public URL (if supported)
	GetPresignedURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// Copy copies an object within storage
	Copy(ctx context.Context, sourcePath, destPath string) (*StorageObject, error)

	// Health checks the health of the storage service
	Health(ctx context.Context) error
}
