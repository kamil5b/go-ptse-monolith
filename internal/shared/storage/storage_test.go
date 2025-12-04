package storage_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage"
	"github.com/kamil5b/go-ptse-monolith/internal/shared/storage/mocks"
)

func TestStorageObjectCreation(t *testing.T) {
	obj := &storage.StorageObject{
		Name:         "file.txt",
		Size:         1024,
		ContentType:  "text/plain",
		ETag:         "abc123",
		LastModified: time.Now(),
		Metadata: map[string]string{
			"key": "value",
		},
	}

	require.NotNil(t, obj)
	assert.Equal(t, "file.txt", obj.Name)
	assert.Equal(t, int64(1024), obj.Size)
	assert.Equal(t, "text/plain", obj.ContentType)
	assert.Equal(t, "abc123", obj.ETag)
	assert.NotNil(t, obj.Metadata)
	assert.Equal(t, "value", obj.Metadata["key"])
}

func TestStorageObjectWithPresignedURL(t *testing.T) {
	presignedURL := "https://example.com/presigned/abc123"
	obj := &storage.StorageObject{
		Name:         "file.txt",
		Size:         1024,
		PresignedURL: presignedURL,
	}

	assert.Equal(t, presignedURL, obj.PresignedURL)
}

func TestUploadOptionsDefaults(t *testing.T) {
	opts := &storage.UploadOptions{
		ContentType: "application/json",
		Metadata: map[string]string{
			"author": "test",
		},
	}

	assert.Equal(t, "application/json", opts.ContentType)
	assert.NotNil(t, opts.Metadata)
	assert.Equal(t, "test", opts.Metadata["author"])
}

func TestUploadOptionsWithAllFields(t *testing.T) {
	opts := &storage.UploadOptions{
		ContentType:          "text/plain",
		CacheControl:         "max-age=3600",
		ContentEncoding:      "gzip",
		ACL:                  "private",
		ServerSideEncryption: true,
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	assert.Equal(t, "text/plain", opts.ContentType)
	assert.Equal(t, "max-age=3600", opts.CacheControl)
	assert.Equal(t, "gzip", opts.ContentEncoding)
	assert.Equal(t, "private", opts.ACL)
	assert.True(t, opts.ServerSideEncryption)
	assert.Equal(t, 2, len(opts.Metadata))
}

func TestMockStorageServiceUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	data := []byte("test content")
	opts := &storage.UploadOptions{ContentType: "text/plain"}
	expectedObj := &storage.StorageObject{
		Name:        "test.txt",
		Size:        int64(len(data)),
		ContentType: "text/plain",
	}

	mockSvc.EXPECT().
		Upload(ctx, "test.txt", gomock.Any(), opts).
		Return(expectedObj, nil).
		Times(1)

	reader := io.NopCloser(strings.NewReader(string(data)))
	obj, err := mockSvc.Upload(ctx, "test.txt", reader, opts)

	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, "test.txt", obj.Name)
}

func TestMockStorageServiceUploadBytes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	data := []byte("test content")
	opts := &storage.UploadOptions{ContentType: "text/plain"}
	expectedObj := &storage.StorageObject{
		Name:        "test.txt",
		Size:        int64(len(data)),
		ContentType: "text/plain",
	}

	mockSvc.EXPECT().
		UploadBytes(ctx, "test.txt", data, opts).
		Return(expectedObj, nil).
		Times(1)

	obj, err := mockSvc.UploadBytes(ctx, "test.txt", data, opts)

	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, "test.txt", obj.Name)
	assert.Equal(t, int64(len(data)), obj.Size)
	assert.Equal(t, "text/plain", obj.ContentType)
}

func TestMockStorageServiceExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	mockSvc.EXPECT().
		Exists(ctx, "test.txt").
		Return(true, nil).
		Times(1)

	exists, err := mockSvc.Exists(ctx, "test.txt")
	require.NoError(t, err)
	assert.True(t, exists)

	mockSvc.EXPECT().
		Exists(ctx, "notfound.txt").
		Return(false, nil).
		Times(1)

	exists, err = mockSvc.Exists(ctx, "notfound.txt")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMockStorageServiceDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	mockSvc.EXPECT().
		Delete(ctx, "test.txt").
		Return(nil).
		Times(1)

	err := mockSvc.Delete(ctx, "test.txt")
	require.NoError(t, err)
}

func TestMockStorageServiceDeletePrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	mockSvc.EXPECT().
		DeletePrefix(ctx, "prefix/").
		Return(nil).
		Times(1)

	err := mockSvc.DeletePrefix(ctx, "prefix/")
	require.NoError(t, err)
}

func TestMockStorageServiceListObjects(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	expectedObjects := []*storage.StorageObject{
		{Name: "docs/file1.txt", Size: 100},
		{Name: "docs/file2.txt", Size: 200},
	}

	mockSvc.EXPECT().
		ListObjects(ctx, "docs/", true).
		Return(expectedObjects, nil).
		Times(1)

	objects, err := mockSvc.ListObjects(ctx, "docs/", true)
	require.NoError(t, err)
	assert.Equal(t, 2, len(objects))
}

func TestMockStorageServiceGetPresignedURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	expectedURL := "https://example.com/presigned/file.txt"

	mockSvc.EXPECT().
		GetPresignedURL(ctx, "file.txt", 1*time.Hour).
		Return(expectedURL, nil).
		Times(1)

	url, err := mockSvc.GetPresignedURL(ctx, "file.txt", 1*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, expectedURL, url)
}

func TestMockStorageServiceHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	ctx := context.Background()

	mockSvc.EXPECT().
		Health(ctx).
		Return(nil).
		Times(1)

	err := mockSvc.Health(ctx)
	require.NoError(t, err)
}

func TestStorageServiceInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockStorageService(ctrl)
	var _ storage.StorageService = mockSvc
}
