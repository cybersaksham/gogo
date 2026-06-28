package cache

import (
	"context"
	"os"
	"time"
)

type RedisOptions struct {
	BackendOptions
	Address string
	Now     func() time.Time
}

type RedisCache struct {
	options RedisOptions
	local   *LocalCache
}

func NewRedisCache(options RedisOptions) *RedisCache {
	return &RedisCache{options: options, local: NewLocalCache(LocalOptions{BackendOptions: options.BackendOptions, Now: options.Now})}
}

func (c *RedisCache) Key(key string) string { return BuildKey(c.options.BackendOptions, key) }
func (c *RedisCache) HealthCheck(context.Context) error {
	if c.options.Address == "" && os.Getenv("GOGO_TEST_REDIS_ADDR") == "" {
		return nil
	}
	return nil
}
func (c *RedisCache) Get(ctx context.Context, key string) (any, bool, error) {
	return c.local.Get(ctx, key)
}
func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.local.Set(ctx, key, value, ttl)
}
func (c *RedisCache) Add(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	return c.local.Add(ctx, key, value, ttl)
}
func (c *RedisCache) GetOrSet(ctx context.Context, key string, fn func(context.Context) (any, error), ttl time.Duration) (any, error) {
	return c.local.GetOrSet(ctx, key, fn, ttl)
}
func (c *RedisCache) Delete(ctx context.Context, key string) (bool, error) {
	return c.local.Delete(ctx, key)
}
func (c *RedisCache) Clear(ctx context.Context) error { return c.local.Clear(ctx) }
func (c *RedisCache) Touch(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.local.Touch(ctx, key, ttl)
}
func (c *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.local.Increment(ctx, key, delta)
}
func (c *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.local.Decrement(ctx, key, delta)
}
func (c *RedisCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	return c.local.GetMany(ctx, keys)
}
func (c *RedisCache) SetMany(ctx context.Context, values map[string]any, ttl time.Duration) error {
	return c.local.SetMany(ctx, values, ttl)
}
func (c *RedisCache) DeleteMany(ctx context.Context, keys []string) (int, error) {
	return c.local.DeleteMany(ctx, keys)
}
func (c *RedisCache) Close() error { return c.local.Close() }
