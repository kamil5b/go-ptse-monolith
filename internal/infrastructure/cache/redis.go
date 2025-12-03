package cache

import (
	"context"
	"fmt"
	"time"

	sharedcache "go-modular-monolith/internal/shared/cache"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	// Connection pool settings
	MaxRetries     int
	PoolSize       int
	MinIdleConns   int
	MaxConnAge     time.Duration
	PoolTimeout    time.Duration
	IdleTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ContextTimeout time.Duration
}

// NewRedisClient creates a new Redis client with the given configuration
func NewRedisClient(cfg RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		PoolTimeout:     cfg.PoolTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxIdleTime: 5 * time.Minute,
	})

	// Test connection with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	return client, nil
}

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// Get retrieves a value from Redis
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", sharedcache.ErrCacheKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("redis get error: %w", err)
	}
	return val, nil
}

// GetBytes retrieves a value from Redis as bytes
func (r *RedisCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, sharedcache.ErrCacheKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("redis get bytes error: %w", err)
	}
	return val, nil
}

// Set stores a value in Redis with optional expiration
func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// SetNX stores a value in Redis only if the key does not exist
func (r *RedisCache) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}
	return ok, nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// Exists checks if one or more keys exist in Redis
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis exists error: %w", err)
	}
	return count, nil
}

// Expire sets a timeout on a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("redis expire error: %w", err)
	}
	return nil
}

// TTL returns the remaining time to live of a key in seconds (-2 if key doesn't exist, -1 if no expiration)
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}
	return ttl, nil
}

// Increment increments the number stored at key by increment
func (r *RedisCache) Increment(ctx context.Context, key string, increment int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, increment).Result()
	if err != nil {
		return 0, fmt.Errorf("redis increment error: %w", err)
	}
	return val, nil
}

// Decrement decrements the number stored at key by decrement
func (r *RedisCache) Decrement(ctx context.Context, key string, decrement int64) (int64, error) {
	val, err := r.client.DecrBy(ctx, key, decrement).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decrement error: %w", err)
	}
	return val, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("redis close error: %w", err)
	}
	return nil
}

// Health checks the health of the Redis connection
func (r *RedisCache) Health(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}
