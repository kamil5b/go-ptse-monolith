package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a string value from cache
	Get(ctx context.Context, key string) (string, error)

	// GetBytes retrieves bytes value from cache
	GetBytes(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with optional expiration
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// SetNX stores a value only if key does not exist
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)

	// Delete removes one or more keys from cache
	Delete(ctx context.Context, keys ...string) error

	// Exists checks if one or more keys exist in cache (returns count of existing keys)
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets a timeout on a key
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// TTL returns the remaining time to live of a key (-2 if key doesn't exist, -1 if no expiration)
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Increment increments the number stored at key
	Increment(ctx context.Context, key string, increment int64) (int64, error)

	// Decrement decrements the number stored at key
	Decrement(ctx context.Context, key string, decrement int64) (int64, error)

	// Health checks the health of the cache
	Health(ctx context.Context) error
}
