package cache

import (
	"context"
	"time"
)

type DummyCache struct{}

func NewDummyCache() *DummyCache { return &DummyCache{} }

func (c *DummyCache) Get(context.Context, string) (any, bool, error)        { return nil, false, nil }
func (c *DummyCache) Set(context.Context, string, any, time.Duration) error { return nil }
func (c *DummyCache) Add(context.Context, string, any, time.Duration) (bool, error) {
	return true, nil
}
func (c *DummyCache) GetOrSet(ctx context.Context, _ string, fn func(context.Context) (any, error), _ time.Duration) (any, error) {
	if fn == nil {
		return nil, ErrInvalidCacheValue
	}
	return fn(ctx)
}
func (c *DummyCache) Delete(context.Context, string) (bool, error) { return false, nil }
func (c *DummyCache) Clear(context.Context) error                  { return nil }
func (c *DummyCache) Touch(context.Context, string, time.Duration) (bool, error) {
	return false, nil
}
func (c *DummyCache) Increment(context.Context, string, int64) (int64, error) {
	return 0, ErrKeyNotFound
}
func (c *DummyCache) Decrement(context.Context, string, int64) (int64, error) {
	return 0, ErrKeyNotFound
}
func (c *DummyCache) GetMany(context.Context, []string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (c *DummyCache) SetMany(context.Context, map[string]any, time.Duration) error { return nil }
func (c *DummyCache) DeleteMany(context.Context, []string) (int, error)            { return 0, nil }
func (c *DummyCache) Close() error                                                 { return nil }
