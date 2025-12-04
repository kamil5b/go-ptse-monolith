package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	storagepkg "github.com/kamil5b/go-ptse-monolith/internal/shared/storage"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GCSStorageConfig configures Google Cloud Storage
type GCSStorageConfig struct {
	ProjectID       string // GCP project ID
	Bucket          string // GCS bucket name
	CredentialsFile string // Path to service account JSON
	CredentialsJSON string // Inline JSON credentials
	StorageClass    string // Storage class (STANDARD, NEARLINE, COLDLINE, ARCHIVE)
	Location        string // Bucket location
	MetadataCache   bool   // Cache object metadata
}

// GCSStorageService stores files in Google Cloud Storage
type GCSStorageService struct {
	config GCSStorageConfig
	client *storage.Client
	bucket *storage.BucketHandle
}

// NewGCSStorageService creates a new GCS storage service
func NewGCSStorageService(ctx context.Context, cfg GCSStorageConfig) (*GCSStorageService, error) {
	var client *storage.Client
	var err error

	// Create GCS client with credentials
	if cfg.CredentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(cfg.CredentialsFile))
	} else if cfg.CredentialsJSON != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorageService{
		config: cfg,
		client: client,
		bucket: client.Bucket(cfg.Bucket),
	}, nil
}

// Upload stores a file from a reader
func (s *GCSStorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storagepkg.UploadOptions,
) (*storagepkg.StorageObject, error) {
	if opts == nil {
		opts = &storagepkg.UploadOptions{}
	}

	object := s.bucket.Object(path)
	writer := object.NewWriter(ctx)

	// Set content type
	if opts.ContentType != "" {
		writer.ContentType = opts.ContentType
	}

	// Set metadata
	if opts.Metadata != nil {
		writer.Metadata = opts.Metadata
	}

	// Set cache control
	if opts.CacheControl != "" {
		writer.CacheControl = opts.CacheControl
	}

	// Set content encoding
	if opts.ContentEncoding != "" {
		writer.ContentEncoding = opts.ContentEncoding
	}

	// Copy data
	if _, err := io.Copy(writer, reader); err != nil {
		return nil, storagepkg.ServiceError("failed to write to GCS", err)
	}

	// Close writer (this actually uploads)
	if err := writer.Close(); err != nil {
		return nil, storagepkg.ServiceError("failed to close writer", err)
	}

	// Get object attributes
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return nil, storagepkg.ServiceError("failed to get object attributes", err)
	}

	return &storagepkg.StorageObject{
		Name:         path,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         attrs.Etag,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
	}, nil
}

// UploadBytes stores bytes
func (s *GCSStorageService) UploadBytes(
	ctx context.Context,
	path string,
	data []byte,
	opts *storagepkg.UploadOptions,
) (*storagepkg.StorageObject, error) {
	reader := bytes.NewReader(data)
	return s.Upload(ctx, path, reader, opts)
}

// Download retrieves a file
func (s *GCSStorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	reader, err := s.bucket.Object(path).NewReader(ctx)
	if err != nil {
		return nil, storagepkg.ServiceError("failed to create reader", err)
	}
	return reader, nil
}

// GetBytes retrieves file contents as bytes
func (s *GCSStorageService) GetBytes(ctx context.Context, path string) ([]byte, error) {
	reader, err := s.Download(ctx, path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, storagepkg.ServiceError("failed to read object", err)
	}

	return buf.Bytes(), nil
}

// GetObject retrieves object metadata
func (s *GCSStorageService) GetObject(ctx context.Context, path string) (*storagepkg.StorageObject, error) {
	attrs, err := s.bucket.Object(path).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, storagepkg.NotFound(path)
		}
		return nil, storagepkg.ServiceError("failed to get object", err)
	}

	return &storagepkg.StorageObject{
		Name:         attrs.Name,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         attrs.Etag,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
	}, nil
}

// Delete removes a file
func (s *GCSStorageService) Delete(ctx context.Context, path string) error {
	err := s.bucket.Object(path).Delete(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return storagepkg.ServiceError("failed to delete object", err)
	}
	return nil
}

// DeletePrefix removes all objects with given prefix
func (s *GCSStorageService) DeletePrefix(ctx context.Context, prefix string) error {
	it := s.bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist {
			break
		}
		if err != nil {
			return storagepkg.ServiceError("failed to list objects", err)
		}

		if err := s.bucket.Object(attrs.Name).Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
			return storagepkg.ServiceError("failed to delete object", err)
		}
	}

	return nil
}

// Exists checks if object exists
func (s *GCSStorageService) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.bucket.Object(path).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, storagepkg.ServiceError("failed to check if object exists", err)
	}
	return true, nil
}

// ListObjects lists objects in a path prefix
func (s *GCSStorageService) ListObjects(
	ctx context.Context,
	prefix string,
	recursive bool,
) ([]*storagepkg.StorageObject, error) {
	query := &storage.Query{Prefix: prefix}
	if !recursive {
		query.Delimiter = "/"
	}

	it := s.bucket.Objects(ctx, query)
	var objects []*storagepkg.StorageObject

	for {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist {
			break
		}
		if err != nil {
			return nil, storagepkg.ServiceError("failed to list objects", err)
		}

		objects = append(objects, &storagepkg.StorageObject{
			Name:         attrs.Name,
			Size:         attrs.Size,
			ContentType:  attrs.ContentType,
			ETag:         attrs.Etag,
			LastModified: attrs.Updated,
			Metadata:     attrs.Metadata,
		})
	}

	return objects, nil
}

// GetPresignedURL generates a temporary public URL
func (s *GCSStorageService) GetPresignedURL(
	ctx context.Context,
	path string,
	expiration time.Duration,
) (string, error) {
	// Generate presigned URL using service account signing
	// This requires a service account with appropriate permissions
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	// For this to work, ensure:
	// 1. Service account key is available (from CredentialsFile or CredentialsJSON)
	// 2. Service account has "iam.serviceAccountUser" and "iam.serviceAccountTokenCreator" roles
	// 3. The bucket is not publicly accessible (enforce this via bucket policy)
	url, err := storage.SignedURL(s.config.Bucket, path, opts)
	if err != nil {
		return "", storagepkg.ServiceError("failed to generate presigned URL", err)
	}

	return url, nil
}

// Copy copies an object within storage
func (s *GCSStorageService) Copy(
	ctx context.Context,
	sourcePath, destPath string,
) (*storagepkg.StorageObject, error) {
	// Copy object
	copier := s.bucket.Object(destPath).CopierFrom(s.bucket.Object(sourcePath))
	attrs, err := copier.Run(ctx)
	if err != nil {
		return nil, storagepkg.ServiceError("failed to copy object", err)
	}

	return &storagepkg.StorageObject{
		Name:         attrs.Name,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         attrs.Etag,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
	}, nil
}

// Health checks the health of the storage service
func (s *GCSStorageService) Health(ctx context.Context) error {
	// Try to access bucket
	attrs, err := s.bucket.Attrs(ctx)
	if err != nil {
		return storagepkg.ServiceError("failed to access GCS bucket", err)
	}

	if attrs == nil {
		return storagepkg.ServiceError("bucket attributes not available", fmt.Errorf("attrs is nil"))
	}

	return nil
}
