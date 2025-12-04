package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3StorageConfig configures the AWS S3 storage
type S3StorageConfig struct {
	Region               string // AWS region
	Bucket               string // S3 bucket name
	AccessKeyID          string // AWS access key
	SecretAccessKey      string // AWS secret key
	Endpoint             string // Custom endpoint (for MinIO, etc)
	UseSSL               bool   // Use SSL for endpoint
	PathStyle            bool   // Use path-style URLs (true for MinIO)
	PresignedURLTTL      int    // Presigned URL validity in seconds
	ServerSideEncryption bool   // Enable encryption
	StorageClass         string // Storage class (STANDARD, GLACIER, etc)
}

// S3StorageService stores files in AWS S3 or compatible services
type S3StorageService struct {
	config     S3StorageConfig
	client     *s3.Client
	presigner  *s3.PresignClient
	uploader   *manager.Uploader
	downloader *manager.Downloader
}

// NewS3StorageService creates a new S3 storage service
func NewS3StorageService(cfg S3StorageConfig) (*S3StorageService, error) {
	ctx := context.Background()

	// Build config options
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.PathStyle
		}
	})

	return &S3StorageService{
		config:     cfg,
		client:     client,
		presigner:  s3.NewPresignClient(client),
		uploader:   manager.NewUploader(client),
		downloader: manager.NewDownloader(client),
	}, nil
}

// Upload stores a file from a reader
func (s *S3StorageService) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	if opts == nil {
		opts = &storage.UploadOptions{}
	}

	// Prepare put object input
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
		Body:   reader,
	}

	// Set content type
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	// Set cache control
	if opts.CacheControl != "" {
		input.CacheControl = aws.String(opts.CacheControl)
	}

	// Set content encoding
	if opts.ContentEncoding != "" {
		input.ContentEncoding = aws.String(opts.ContentEncoding)
	}

	// Set metadata
	if opts.Metadata != nil {
		input.Metadata = opts.Metadata
	}

	// Set ACL
	if opts.ACL != "" {
		input.ACL = types.ObjectCannedACL(opts.ACL)
	}

	// Set storage class
	if s.config.StorageClass != "" {
		input.StorageClass = types.StorageClass(s.config.StorageClass)
	}

	// Set server-side encryption
	if s.config.ServerSideEncryption {
		input.ServerSideEncryption = types.ServerSideEncryptionAes256
	}

	// Upload file
	output, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, storage.ServiceError("failed to upload object", err)
	}

	// Get object metadata
	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, storage.ServiceError("failed to get object metadata", err)
	}

	// Generate presigned URL
	presignInput := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	}

	presignOutput, err := s.presigner.PresignGetObject(ctx, presignInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(s.config.PresignedURLTTL) * time.Second
		},
	)
	if err != nil {
		return nil, storage.ServiceError("failed to generate presigned URL", err)
	}

	result := &storage.StorageObject{
		Name:         path,
		Size:         aws.ToInt64(headOutput.ContentLength),
		ContentType:  aws.ToString(headOutput.ContentType),
		ETag:         aws.ToString(output.ETag),
		LastModified: aws.ToTime(headOutput.LastModified),
		Metadata:     headOutput.Metadata,
		PresignedURL: presignOutput.URL,
	}

	return result, nil
}

// UploadBytes stores bytes
func (s *S3StorageService) UploadBytes(
	ctx context.Context,
	path string,
	data []byte,
	opts *storage.UploadOptions,
) (*storage.StorageObject, error) {
	reader := bytes.NewReader(data)
	return s.Upload(ctx, path, reader, opts)
}

// Download retrieves a file
func (s *S3StorageService) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, storage.ServiceError("failed to download object", err)
	}

	return output.Body, nil
}

// GetBytes retrieves file contents as bytes
func (s *S3StorageService) GetBytes(ctx context.Context, path string) ([]byte, error) {
	reader, err := s.Download(ctx, path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, storage.ServiceError("failed to read object", err)
	}

	return buf.Bytes(), nil
}

// GetObject retrieves object metadata
func (s *S3StorageService) GetObject(ctx context.Context, path string) (*storage.StorageObject, error) {
	output, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, storage.ServiceError("failed to get object metadata", err)
	}

	return &storage.StorageObject{
		Name:         path,
		Size:         aws.ToInt64(output.ContentLength),
		ContentType:  aws.ToString(output.ContentType),
		ETag:         aws.ToString(output.ETag),
		LastModified: aws.ToTime(output.LastModified),
		Metadata:     output.Metadata,
	}, nil
}

// Delete removes a file
func (s *S3StorageService) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return storage.ServiceError("failed to delete object", err)
	}

	return nil
}

// DeletePrefix removes all objects with given prefix
func (s *S3StorageService) DeletePrefix(ctx context.Context, prefix string) error {
	// List objects with prefix
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return storage.ServiceError("failed to list objects", err)
		}

		// Delete objects
		if len(page.Contents) > 0 {
			objectIdentifiers := make([]types.ObjectIdentifier, len(page.Contents))
			for i, obj := range page.Contents {
				objectIdentifiers[i] = types.ObjectIdentifier{
					Key: obj.Key,
				}
			}

			_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(s.config.Bucket),
				Delete: &types.Delete{
					Objects: objectIdentifiers,
				},
			})
			if err != nil {
				return storage.ServiceError("failed to delete objects", err)
			}
		}
	}

	return nil
}

// Exists checks if object exists
func (s *S3StorageService) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		// Check if it's a "not found" error
		var nfe *types.NotFound
		if errors.As(err, &nfe) {
			return false, nil
		}
		return false, storage.ServiceError("failed to check if object exists", err)
	}

	return true, nil
}

// ListObjects lists objects in a path prefix
func (s *S3StorageService) ListObjects(
	ctx context.Context,
	prefix string,
	recursive bool,
) ([]*storage.StorageObject, error) {
	var objects []*storage.StorageObject

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(prefix),
	})

	// Set delimiter if not recursive
	if !recursive {
		paginator = s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
			Bucket:    aws.String(s.config.Bucket),
			Prefix:    aws.String(prefix),
			Delimiter: aws.String("/"),
		})
	}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, storage.ServiceError("failed to list objects", err)
		}

		if page.Contents != nil {
			for _, obj := range page.Contents {
				objects = append(objects, &storage.StorageObject{
					Name:         aws.ToString(obj.Key),
					Size:         aws.ToInt64(obj.Size),
					ETag:         aws.ToString(obj.ETag),
					LastModified: aws.ToTime(obj.LastModified),
					Metadata:     make(map[string]string),
				})
			}
		}
	}

	return objects, nil
}

// GetPresignedURL generates a temporary public URL
func (s *S3StorageService) GetPresignedURL(
	ctx context.Context,
	path string,
	expiration time.Duration,
) (string, error) {
	presignInput := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(path),
	}

	presignOutput, err := s.presigner.PresignGetObject(ctx, presignInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = expiration
		},
	)
	if err != nil {
		return "", storage.ServiceError("failed to generate presigned URL", err)
	}

	return presignOutput.URL, nil
}

// Copy copies an object within storage
func (s *S3StorageService) Copy(
	ctx context.Context,
	sourcePath, destPath string,
) (*storage.StorageObject, error) {
	// Copy object
	output, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.Bucket),
		CopySource: aws.String(s.config.Bucket + "/" + sourcePath),
		Key:        aws.String(destPath),
	})
	if err != nil {
		return nil, storage.ServiceError("failed to copy object", err)
	}

	// Get destination object metadata
	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(destPath),
	})
	if err != nil {
		return nil, storage.ServiceError("failed to get copied object metadata", err)
	}

	result := &storage.StorageObject{
		Name:         destPath,
		Size:         aws.ToInt64(headOutput.ContentLength),
		ContentType:  aws.ToString(headOutput.ContentType),
		ETag:         aws.ToString(output.CopyObjectResult.ETag),
		LastModified: aws.ToTime(headOutput.LastModified),
		Metadata:     headOutput.Metadata,
	}

	return result, nil
}

// Health checks the health of the storage service
func (s *S3StorageService) Health(ctx context.Context) error {
	// Try to list objects in the bucket
	_, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.config.Bucket),
		MaxKeys: aws.Int32(1),
	})

	if err != nil {
		return storage.ServiceError("failed to access S3 bucket", err)
	}

	return nil
}
