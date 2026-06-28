package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LocalOptions struct {
	BackendOptions
	MaxEntries int
	Now        func() time.Time
}

type LocalCache struct {
	mu       sync.RWMutex
	options  LocalOptions
	now      func() time.Time
	items    map[string]localItem
	inserted []string
	closed   bool
}

type localItem struct {
	value     any
	expiresAt time.Time
}

func NewLocalCache(options LocalOptions) *LocalCache {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &LocalCache{options: options, now: now, items: map[string]localItem{}}
}

func (c *LocalCache) Get(ctx context.Context, key string) (any, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return nil, false, err
	}
	item, ok := c.items[c.key(key)]
	if !ok {
		return nil, false, nil
	}
	if c.expired(item) {
		delete(c.items, c.key(key))
		return nil, false, nil
	}
	return item.value, true, nil
}

func (c *LocalCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return err
	}
	c.setLocked(key, value, ttl)
	return nil
}

func (c *LocalCache) Add(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return false, err
	}
	internal := c.key(key)
	if item, ok := c.items[internal]; ok && !c.expired(item) {
		return false, nil
	}
	c.setLocked(key, value, ttl)
	return true, nil
}

func (c *LocalCache) GetOrSet(ctx context.Context, key string, fn func(context.Context) (any, error), ttl time.Duration) (any, error) {
	if value, ok, err := c.Get(ctx, key); err != nil || ok {
		return value, err
	}
	if fn == nil {
		return nil, fmt.Errorf("%w: get-or-set function is required", ErrInvalidCacheValue)
	}
	value, err := fn(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.Set(ctx, key, value, ttl); err != nil {
		return nil, err
	}
	return value, nil
}

func (c *LocalCache) Delete(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return false, err
	}
	internal := c.key(key)
	_, ok := c.items[internal]
	delete(c.items, internal)
	return ok, nil
}

func (c *LocalCache) Clear(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return err
	}
	c.items = map[string]localItem{}
	c.inserted = nil
	return nil
}

func (c *LocalCache) Touch(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return false, err
	}
	internal := c.key(key)
	item, ok := c.items[internal]
	if !ok || c.expired(item) {
		delete(c.items, internal)
		return false, nil
	}
	item.expiresAt = expiry(c.now(), ttl)
	c.items[internal] = item
	return true, nil
}

func (c *LocalCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.addInt(ctx, key, delta)
}

func (c *LocalCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.addInt(ctx, key, -delta)
}

func (c *LocalCache) GetMany(ctx context.Context, keys []string) (map[string]any, error) {
	values := make(map[string]any, len(keys))
	for _, key := range keys {
		value, ok, err := c.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if ok {
			values[key] = value
		}
	}
	return values, nil
}

func (c *LocalCache) SetMany(ctx context.Context, values map[string]any, ttl time.Duration) error {
	for key, value := range values {
		if err := c.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

func (c *LocalCache) DeleteMany(ctx context.Context, keys []string) (int, error) {
	count := 0
	for _, key := range keys {
		deleted, err := c.Delete(ctx, key)
		if err != nil {
			return count, err
		}
		if deleted {
			count++
		}
	}
	return count, nil
}

func (c *LocalCache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	for key, item := range c.items {
		if c.expired(item) {
			delete(c.items, key)
			count++
		}
	}
	return count
}

func (c *LocalCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	c.items = nil
	c.inserted = nil
	return nil
}

func (c *LocalCache) addInt(ctx context.Context, key string, delta int64) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureOpen(); err != nil {
		return 0, err
	}
	internal := c.key(key)
	item, ok := c.items[internal]
	if !ok || c.expired(item) {
		return 0, ErrKeyNotFound
	}
	current, err := asInt64(item.value)
	if err != nil {
		return 0, err
	}
	current += delta
	item.value = current
	c.items[internal] = item
	return current, nil
}

func (c *LocalCache) setLocked(key string, value any, ttl time.Duration) {
	internal := c.key(key)
	if _, exists := c.items[internal]; !exists {
		c.inserted = append(c.inserted, internal)
	}
	c.items[internal] = localItem{value: value, expiresAt: expiry(c.now(), ttl)}
	c.enforceMaxEntries()
}

func (c *LocalCache) enforceMaxEntries() {
	if c.options.MaxEntries <= 0 {
		return
	}
	for len(c.items) > c.options.MaxEntries && len(c.inserted) > 0 {
		oldest := c.inserted[0]
		c.inserted = c.inserted[1:]
		if _, ok := c.items[oldest]; ok {
			delete(c.items, oldest)
		}
	}
}

func (c *LocalCache) key(key string) string {
	return BuildKey(c.options.BackendOptions, key)
}

func (c *LocalCache) expired(item localItem) bool {
	return !item.expiresAt.IsZero() && !c.now().Before(item.expiresAt)
}

func (c *LocalCache) ensureOpen() error {
	if c.closed {
		return ErrBackendClosed
	}
	return nil
}

func expiry(now time.Time, ttl time.Duration) time.Time {
	if ttl == 0 {
		return time.Time{}
	}
	return now.Add(ttl)
}
