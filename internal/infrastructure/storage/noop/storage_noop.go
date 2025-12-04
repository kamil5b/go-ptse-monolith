package noop

import (
	"context"
	"io"
	"time"

	"go-modular-monolith/internal/shared/storage"
)

// NoOpStorageService is a no-op implementation of StorageService for testing and development
type NoOpStorageService struct{}

// NewNoOpStorageService creates a new NoOp storage service
func NewNoOpStorageService() *NoOpStorageService {
	return &NoOpStorageService{}
}

// Upload is a no-op implementation
func (s *NoOpStorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	return &storage.StorageObject{
		Name:        path,
		Size:        0,
		ContentType: opts.ContentType,
		ETag:        "noop-etag",
		Metadata:    opts.Metadata,
	}, nil
}

// UploadBytes is a no-op implementation
func (s *NoOpStorageService) UploadBytes(
	ctx context.Context,
	path string,
	data []byte,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	return &storage.StorageObject{
		Name:        path,
		Size:        int64(len(data)),
		ContentType: opts.ContentType,
		ETag:        "noop-etag",
		Metadata:    opts.Metadata,
	}, nil
}

// Download is a no-op implementation
func (s *NoOpStorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}

// GetBytes is a no-op implementation
func (s *NoOpStorageService) GetBytes(ctx context.Context, path string) ([]byte, error) {
	return []byte{}, nil
}

// GetObject is a no-op implementation
func (s *NoOpStorageService) GetObject(ctx context.Context, path string) (*storage.StorageObject, error) {
	return &storage.StorageObject{
		Name:         path,
		Size:         0,
		ETag:         "noop-etag",
		LastModified: time.Now(),
	}, nil
}

// Delete is a no-op implementation
func (s *NoOpStorageService) Delete(ctx context.Context, path string) error {
	return nil
}

// DeletePrefix is a no-op implementation
func (s *NoOpStorageService) DeletePrefix(ctx context.Context, prefix string) error {
	return nil
}

// Exists is a no-op implementation
func (s *NoOpStorageService) Exists(ctx context.Context, path string) (bool, error) {
	return true, nil
}

// ListObjects is a no-op implementation
func (s *NoOpStorageService) ListObjects(
	ctx context.Context,
	prefix string,
	recursive bool,
) ([]*storage.StorageObject, error) {
	return []*storage.StorageObject{}, nil
}

// GetPresignedURL is a no-op implementation
func (s *NoOpStorageService) GetPresignedURL(
	ctx context.Context,
	path string,
	expiration time.Duration,
) (string, error) {
	return "http://example.com/" + path, nil
}

// Copy is a no-op implementation
func (s *NoOpStorageService) Copy(
	ctx context.Context,
	sourcePath, destPath string,
) (*storage.StorageObject, error) {
	return &storage.StorageObject{
		Name:         destPath,
		Size:         0,
		ETag:         "noop-etag",
		LastModified: time.Now(),
	}, nil
}

// Health is a no-op implementation
func (s *NoOpStorageService) Health(ctx context.Context) error {
	return nil
}
