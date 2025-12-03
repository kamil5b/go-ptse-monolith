package cache

import "errors"

var (
	// ErrCacheKeyNotFound is returned when a key is not found in cache
	ErrCacheKeyNotFound = errors.New("cache key not found")

	// ErrCacheDisabled is returned when cache operations are called but cache is disabled
	ErrCacheDisabled = errors.New("cache is disabled")
)
