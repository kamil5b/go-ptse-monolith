package cache

import (
	"context"
	"sync"
	"time"
)

// cacheEntry holds a cached value with expiration time
type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
	isExpired bool
}

// InMemoryCache is a simple in-memory cache implementation for testing
type InMemoryCache struct {
	data map[string]*cacheEntry
	mu   sync.RWMutex
}

// NewInMemoryCache creates a new in-memory cache instance
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		data: make(map[string]*cacheEntry),
	}

	// Start cleanup goroutine to remove expired entries periodically
	go cache.cleanupExpired()

	return cache
}

// cleanupExpired removes expired entries from the cache
func (c *InMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if !entry.isExpired && !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Get retrieves a string value from cache
func (c *InMemoryCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return "", ErrCacheKeyNotFound
	}

	if entry.isExpired || (!entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt)) {
		return "", ErrCacheKeyNotFound
	}

	val, ok := entry.value.(string)
	if !ok {
		return "", ErrCacheKeyNotFound
	}

	return val, nil
}

// GetBytes retrieves bytes value from cache
func (c *InMemoryCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, ErrCacheKeyNotFound
	}

	if entry.isExpired || (!entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt)) {
		return nil, ErrCacheKeyNotFound
	}

	val, ok := entry.value.([]byte)
	if !ok {
		return nil, ErrCacheKeyNotFound
	}

	return val, nil
}

// Set stores a value in cache with optional expiration
func (c *InMemoryCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Time{}
	if expiration > 0 {
		expiresAt = time.Now().Add(expiration)
	}

	c.data[key] = &cacheEntry{
		value:     value,
		expiresAt: expiresAt,
		isExpired: false,
	}

	return nil
}

// SetNX stores a value only if key does not exist
func (c *InMemoryCache) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if exists && !entry.isExpired && (entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt)) {
		return false, nil
	}

	expiresAt := time.Time{}
	if expiration > 0 {
		expiresAt = time.Now().Add(expiration)
	}

	c.data[key] = &cacheEntry{
		value:     value,
		expiresAt: expiresAt,
		isExpired: false,
	}

	return true, nil
}

// Delete removes one or more keys from cache
func (c *InMemoryCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.data, key)
	}

	return nil
}

// Exists checks if one or more keys exist in cache
func (c *InMemoryCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := int64(0)
	now := time.Now()

	for _, key := range keys {
		entry, exists := c.data[key]
		if exists && !entry.isExpired && (entry.expiresAt.IsZero() || now.Before(entry.expiresAt)) {
			count++
		}
	}

	return count, nil
}

// Expire sets a timeout on a key
func (c *InMemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if !exists {
		return ErrCacheKeyNotFound
	}

	entry.expiresAt = time.Now().Add(expiration)
	return nil
}

// TTL returns the remaining time to live of a key
func (c *InMemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return 0, ErrCacheKeyNotFound
	}

	if entry.isExpired || entry.expiresAt.IsZero() {
		return -1, nil // -1 means no expiration
	}

	ttl := time.Until(entry.expiresAt)
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}

// Increment increments the number stored at key
func (c *InMemoryCache) Increment(ctx context.Context, key string, increment int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if !exists {
		c.data[key] = &cacheEntry{value: increment, expiresAt: time.Time{}, isExpired: false}
		return increment, nil
	}

	val, ok := entry.value.(int64)
	if !ok {
		return 0, ErrCacheKeyNotFound
	}

	newVal := val + increment
	entry.value = newVal

	return newVal, nil
}

// Decrement decrements the number stored at key
func (c *InMemoryCache) Decrement(ctx context.Context, key string, decrement int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if !exists {
		newVal := -decrement
		c.data[key] = &cacheEntry{value: newVal, expiresAt: time.Time{}, isExpired: false}
		return newVal, nil
	}

	val, ok := entry.value.(int64)
	if !ok {
		return 0, ErrCacheKeyNotFound
	}

	newVal := val - decrement
	entry.value = newVal

	return newVal, nil
}

// Health checks the health of the cache
func (c *InMemoryCache) Health(ctx context.Context) error {
	return nil // In-memory cache is always healthy
}
