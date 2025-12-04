package local

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
)

// LocalStorageConfig configures the local filesystem storage
type LocalStorageConfig struct {
	BasePath          string      // Base directory path
	MaxFileSize       int64       // Maximum file size in bytes
	AllowPublicAccess bool        // Serve files via HTTP
	PublicURL         string      // Public URL prefix
	CreateMissingDirs bool        // Auto-create directories
	FilePermissions   os.FileMode // File permissions (default: 0644)
	DirPermissions    os.FileMode // Directory permissions (default: 0755)
}

// LocalStorageService stores files in local filesystem
type LocalStorageService struct {
	config LocalStorageConfig
}

// NewLocalStorageService creates a new local storage service
func NewLocalStorageService(config LocalStorageConfig) (*LocalStorageService, error) {
	// Set defaults
	if config.FilePermissions == 0 {
		config.FilePermissions = 0644
	}
	if config.DirPermissions == 0 {
		config.DirPermissions = 0755
	}

	// Validate and create base directory
	basePath := filepath.Clean(config.BasePath)
	if basePath == "" || basePath == "." {
		return nil, fmt.Errorf("invalid base path: %s", config.BasePath)
	}

	if err := os.MkdirAll(basePath, config.DirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorageService{config: config}, nil
}

// Upload stores a file from a reader
func (s *LocalStorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	if opts == nil {
		opts = &storage.UploadOptions{}
	}

	fullPath, err := s.validatePath(path)
	if err != nil {
		return nil, err
	}

	// Create directory if needed
	if s.config.CreateMissingDirs {
		dirPath := filepath.Dir(fullPath)
		if err := os.MkdirAll(dirPath, s.config.DirPermissions); err != nil {
			return nil, storage.ServiceError("failed to create directory", err)
		}
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, storage.ServiceError("failed to create file", err)
	}
	defer file.Close()

	// Copy data with size limit check
	limitedReader := io.LimitReader(reader, s.config.MaxFileSize+1)
	written, err := io.Copy(file, limitedReader)
	if err != nil {
		os.Remove(fullPath)
		return nil, storage.ServiceError("failed to write file", err)
	}

	// Check size limit
	if written > s.config.MaxFileSize {
		os.Remove(fullPath)
		return nil, storage.SizeLimitExceeded(s.config.MaxFileSize)
	}

	// Change permissions
	if err := os.Chmod(fullPath, s.config.FilePermissions); err != nil {
		return nil, storage.ServiceError("failed to set permissions", err)
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, storage.ServiceError("failed to get file info", err)
	}

	// Calculate ETag (MD5)
	fileContent, _ := os.ReadFile(fullPath)
	hash := md5.Sum(fileContent)
	etag := fmt.Sprintf("%x", hash)

	result := &storage.StorageObject{
		Name:         path,
		Size:         written,
		ContentType:  opts.ContentType,
		ETag:         etag,
		LastModified: info.ModTime().UTC(),
		Metadata:     opts.Metadata,
	}

	// Set presigned URL if public access enabled
	if s.config.AllowPublicAccess && s.config.PublicURL != "" {
		result.PresignedURL = s.config.PublicURL + "/" + path
	}

	return result, nil
}

// UploadBytes stores bytes
func (s *LocalStorageService) UploadBytes(
	ctx context.Context,
	path string,
	data []byte,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	reader := bytes.NewReader(data)
	return s.Upload(ctx, path, reader, opts)
}

// Download retrieves a file
func (s *LocalStorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath, err := s.validatePath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NotFound(path)
		}
		return nil, storage.ServiceError("failed to open file", err)
	}

	return file, nil
}

// GetBytes retrieves file contents as bytes
func (s *LocalStorageService) GetBytes(ctx context.Context, path string) ([]byte, error) {
	fullPath, err := s.validatePath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NotFound(path)
		}
		return nil, storage.ServiceError("failed to read file", err)
	}

	return data, nil
}

// GetObject retrieves object metadata
func (s *LocalStorageService) GetObject(ctx context.Context, path string) (*storage.StorageObject, error) {
	fullPath, err := s.validatePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NotFound(path)
		}
		return nil, storage.ServiceError("failed to stat file", err)
	}

	// Calculate ETag
	fileContent, _ := os.ReadFile(fullPath)
	hash := md5.Sum(fileContent)
	etag := fmt.Sprintf("%x", hash)

	result := &storage.StorageObject{
		Name:         path,
		Size:         info.Size(),
		ETag:         etag,
		LastModified: info.ModTime().UTC(),
		Metadata:     make(map[string]string),
	}

	return result, nil
}

// Delete removes a file
func (s *LocalStorageService) Delete(ctx context.Context, path string) error {
	fullPath, err := s.validatePath(path)
	if err != nil {
		return err
	}

	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return storage.ServiceError("failed to delete file", err)
	}

	return nil
}

// DeletePrefix removes all objects with given prefix
func (s *LocalStorageService) DeletePrefix(ctx context.Context, prefix string) error {
	fullPath := filepath.Join(s.config.BasePath, prefix)

	// Validate prefix path
	if !isPathWithinBase(fullPath, s.config.BasePath) {
		return storage.InvalidPath(prefix)
	}

	// Remove directory or file
	err := os.RemoveAll(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return storage.ServiceError("failed to delete prefix", err)
	}

	return nil
}

// Exists checks if object exists
func (s *LocalStorageService) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := s.validatePath(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, err
}

// ListObjects lists objects in a path prefix
func (s *LocalStorageService) ListObjects(
	ctx context.Context,
	prefix string,
	recursive bool,
) ([]*storage.StorageObject, error) {
	fullPrefix := filepath.Join(s.config.BasePath, prefix)

	if !isPathWithinBase(fullPrefix, s.config.BasePath) {
		return nil, storage.InvalidPath(prefix)
	}

	var objects []*storage.StorageObject

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the prefix directory itself
		if path == fullPrefix {
			return nil
		}

		// For non-recursive, skip nested directories
		if !recursive {
			relPath := filepath.Dir(filepath.Clean(path[len(fullPrefix):]))
			if relPath != "." && relPath != "" {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() {
			relPath := path[len(s.config.BasePath):]
			if relPath[0] == os.PathSeparator {
				relPath = relPath[1:]
			}

			// Calculate ETag
			fileContent, _ := os.ReadFile(path)
			hash := md5.Sum(fileContent)
			etag := fmt.Sprintf("%x", hash)

			obj := &storage.StorageObject{
				Name:         filepath.ToSlash(relPath),
				Size:         info.Size(),
				ETag:         etag,
				LastModified: info.ModTime().UTC(),
				Metadata:     make(map[string]string),
			}

			objects = append(objects, obj)
		}

		return nil
	}

	if err := filepath.Walk(fullPrefix, walkFunc); err != nil && !os.IsNotExist(err) {
		return nil, storage.ServiceError("failed to list objects", err)
	}

	return objects, nil
}

// GetPresignedURL generates a temporary public URL
func (s *LocalStorageService) GetPresignedURL(
	ctx context.Context,
	path string,
	expiration time.Duration,
) (string, error) {
	if !s.config.AllowPublicAccess || s.config.PublicURL == "" {
		return "", storage.PermissionDenied("presigned URLs not enabled")
	}

	_, err := s.validatePath(path)
	if err != nil {
		return "", err
	}

	// For local storage, presigned URL is just the public URL
	// (In practice, you'd implement token-based expiration)
	return s.config.PublicURL + "/" + path, nil
}

// Copy copies an object within storage
func (s *LocalStorageService) Copy(
	ctx context.Context,
	sourcePath, destPath string,
) (*storage.StorageObject, error) {
	// Validate both paths
	fullSourcePath, err := s.validatePath(sourcePath)
	if err != nil {
		return nil, err
	}

	fullDestPath, err := s.validatePath(destPath)
	if err != nil {
		return nil, err
	}

	// Create destination directory
	destDir := filepath.Dir(fullDestPath)
	if s.config.CreateMissingDirs {
		if err := os.MkdirAll(destDir, s.config.DirPermissions); err != nil {
			return nil, storage.ServiceError("failed to create destination directory", err)
		}
	}

	// Read source file
	sourceData, err := os.ReadFile(fullSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NotFound(sourcePath)
		}
		return nil, storage.ServiceError("failed to read source file", err)
	}

	// Write destination file
	if err := os.WriteFile(fullDestPath, sourceData, s.config.FilePermissions); err != nil {
		return nil, storage.ServiceError("failed to write destination file", err)
	}

	// Get destination file info
	info, err := os.Stat(fullDestPath)
	if err != nil {
		return nil, storage.ServiceError("failed to stat destination file", err)
	}

	hash := md5.Sum(sourceData)
	etag := fmt.Sprintf("%x", hash)

	result := &storage.StorageObject{
		Name:         destPath,
		Size:         info.Size(),
		ETag:         etag,
		LastModified: info.ModTime().UTC(),
		Metadata:     make(map[string]string),
	}

	return result, nil
}

// Health checks the health of the storage service
func (s *LocalStorageService) Health(ctx context.Context) error {
	// Try to access base directory
	if _, err := os.Stat(s.config.BasePath); err != nil {
		return storage.ServiceError("storage directory not accessible", err)
	}

	// Try to create and delete a test file
	testPath := filepath.Join(s.config.BasePath, ".health-check")
	if err := os.WriteFile(testPath, []byte("ok"), 0644); err != nil {
		return storage.ServiceError("cannot write to storage directory", err)
	}

	if err := os.Remove(testPath); err != nil {
		return storage.ServiceError("cannot delete from storage directory", err)
	}

	return nil
}

// validatePath ensures the path is within the base directory
func (s *LocalStorageService) validatePath(path string) (string, error) {
	if path == "" {
		return "", storage.InvalidPath("empty path")
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Prevent directory traversal
	fullPath := filepath.Join(s.config.BasePath, cleanPath)
	fullPath = filepath.Clean(fullPath)

	if !isPathWithinBase(fullPath, s.config.BasePath) {
		return "", storage.InvalidPath(path)
	}

	return fullPath, nil
}

// isPathWithinBase checks if a path is within the base directory
func isPathWithinBase(path, baseDir string) bool {
	path = filepath.Clean(path)
	baseDir = filepath.Clean(baseDir)

	// Add trailing separator to baseDir to avoid partial matches
	if !filepath.IsAbs(baseDir) {
		baseDir, _ = filepath.Abs(baseDir)
	}
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}

	return path == baseDir || (filepath.HasPrefix(path, baseDir+string(filepath.Separator)))
}
