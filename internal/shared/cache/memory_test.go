package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryCache(t *testing.T) {
	cache := NewInMemoryCache()
	require.NotNil(t, cache)
	assert.NotNil(t, cache.data)
}

func TestInMemoryCacheSet(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)

	val, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestInMemoryCacheSetBytes(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	data := []byte("test data")
	err := cache.Set(ctx, "key1", data, 1*time.Hour)
	require.NoError(t, err)

	retrieved, err := cache.GetBytes(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)
}

func TestInMemoryCacheGetNonExistent(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)
}

func TestInMemoryCacheDelete(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Hour)
	err := cache.Delete(ctx, "key1")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)
}

func TestInMemoryCacheDeleteMultiple(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Hour)
	cache.Set(ctx, "key2", "value2", 1*time.Hour)
	cache.Set(ctx, "key3", "value3", 1*time.Hour)

	err := cache.Delete(ctx, "key1", "key2")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)

	_, err = cache.Get(ctx, "key2")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)

	val, err := cache.Get(ctx, "key3")
	require.NoError(t, err)
	assert.Equal(t, "value3", val)
}

func TestInMemoryCacheSetNX(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// SetNX on non-existent key should succeed
	set, err := cache.SetNX(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)
	assert.True(t, set)

	// SetNX on existing key should fail
	set, err = cache.SetNX(ctx, "key1", "value2", 1*time.Hour)
	require.NoError(t, err)
	assert.False(t, set)

	// Value should not be changed
	val, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestInMemoryCacheExists(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Hour)
	cache.Set(ctx, "key2", "value2", 1*time.Hour)

	count, err := cache.Exists(ctx, "key1", "key2", "key3")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestInMemoryCacheIncrement(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// Increment non-existent key
	val, err := cache.Increment(ctx, "counter", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)

	// Increment existing key
	val, err = cache.Increment(ctx, "counter", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(8), val)
}

func TestInMemoryCacheDecrement(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// Decrement non-existent key
	val, err := cache.Decrement(ctx, "counter", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(-5), val)

	// Decrement existing key
	val, err = cache.Decrement(ctx, "counter", 3)
	require.NoError(t, err)
	assert.Equal(t, int64(-8), val)
}

func TestInMemoryCacheExpiration(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// Set with short expiration
	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	require.NoError(t, err)

	val, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)
}

func TestInMemoryCacheTTL(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// TTL for non-existent key
	_, err := cache.TTL(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)

	// Set with expiration
	cache.Set(ctx, "key1", "value1", 1*time.Hour)
	ttl, err := cache.TTL(ctx, "key1")
	require.NoError(t, err)

	// TTL should be close to 1 hour (allow 5 second variance)
	assert.Greater(t, ttl, 59*time.Minute)
	assert.Less(t, ttl, 61*time.Minute)

	// Set without expiration
	cache.Set(ctx, "key2", "value2", 0)
	ttl, err = cache.TTL(ctx, "key2")
	require.NoError(t, err)
	assert.Equal(t, -1*time.Duration(1), ttl)
}

func TestInMemoryCacheExpire(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Hour)
	err := cache.Expire(ctx, "key1", 100*time.Millisecond)
	require.NoError(t, err)

	// Should still exist
	val, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should now be expired
	_, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)
}

func TestInMemoryCacheExpireNonExistent(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	err := cache.Expire(ctx, "nonexistent", 1*time.Hour)
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)
}

func TestInMemoryCacheHealth(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	err := cache.Health(ctx)
	assert.NoError(t, err)
}

func TestInMemoryCacheTypeConversion(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	// Set a string value and try to get as bytes
	cache.Set(ctx, "key1", "string_value", 1*time.Hour)
	_, err := cache.GetBytes(ctx, "key1")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)

	// Set bytes and try to get as string
	cache.Set(ctx, "key2", []byte("bytes_value"), 1*time.Hour)
	_, err = cache.Get(ctx, "key2")
	assert.ErrorIs(t, err, ErrCacheKeyNotFound)
}

func TestInMemoryCacheDeleteEmpty(t *testing.T) {
	cache := NewInMemoryCache()
	ctx := context.Background()

	err := cache.Delete(ctx)
	assert.NoError(t, err)
}
