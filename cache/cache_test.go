package cache

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMemoryStoreGetSetAndExpiry(t *testing.T) {
	store := NewMemoryStore()
	entry := Entry{Status: 200, Header: http.Header{"X-Test": []string{"yes"}}, Body: []byte("cached")}

	if err := store.Set(context.Background(), "key", entry, time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, ok, err := store.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok || got.Status != 200 || string(got.Body) != "cached" || got.Header.Get("X-Test") != "yes" {
		t.Fatalf("Get() = (%#v, %v), want cached entry", got, ok)
	}

	if err := store.Set(context.Background(), "expired", entry, -time.Second); err != nil {
		t.Fatalf("Set(expired) error = %v", err)
	}
	_, ok, err = store.Get(context.Background(), "expired")
	if err != nil {
		t.Fatalf("Get(expired) error = %v", err)
	}
	if ok {
		t.Fatalf("expired entry was returned")
	}
}

func TestLocalCacheOperationsTTLMaxEntriesAndBulk(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	cache := NewLocalCache(LocalOptions{MaxEntries: 2, Now: func() time.Time { return now }})
	ctx := context.Background()
	if err := cache.Set(ctx, "a", "one", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if added, err := cache.Add(ctx, "a", "two", time.Minute); err != nil || added {
		t.Fatalf("Add(existing) = %v, %v", added, err)
	}
	if value, ok, err := cache.Get(ctx, "a"); err != nil || !ok || value != "one" {
		t.Fatalf("Get(a) = %#v, %v, %v", value, ok, err)
	}
	if value, err := cache.GetOrSet(ctx, "b", func(context.Context) (any, error) { return int64(2), nil }, time.Minute); err != nil || value != int64(2) {
		t.Fatalf("GetOrSet() = %#v, %v", value, err)
	}
	if value, err := cache.Increment(ctx, "b", 3); err != nil || value != 5 {
		t.Fatalf("Increment() = %d, %v", value, err)
	}
	if value, err := cache.Decrement(ctx, "b", 2); err != nil || value != 3 {
		t.Fatalf("Decrement() = %d, %v", value, err)
	}
	if ok, err := cache.Touch(ctx, "a", 2*time.Minute); err != nil || !ok {
		t.Fatalf("Touch() = %v, %v", ok, err)
	}
	if err := cache.SetMany(ctx, map[string]any{"c": "three", "d": "four"}, time.Minute); err != nil {
		t.Fatalf("SetMany() error = %v", err)
	}
	if _, ok, _ := cache.Get(ctx, "a"); ok {
		t.Fatalf("max entries should evict oldest key")
	}
	values, err := cache.GetMany(ctx, []string{"c", "d", "missing"})
	if err != nil || len(values) != 2 || values["c"] != "three" || values["d"] != "four" {
		t.Fatalf("GetMany() = %#v, %v", values, err)
	}
	if deleted, err := cache.DeleteMany(ctx, []string{"c", "d"}); err != nil || deleted != 2 {
		t.Fatalf("DeleteMany() = %d, %v", deleted, err)
	}
	if err := cache.Set(ctx, "expired", "gone", time.Minute); err != nil {
		t.Fatalf("Set(expired) error = %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, ok, _ := cache.Get(ctx, "expired"); ok {
		t.Fatalf("expired generic value was returned")
	}
	if err := cache.Clear(ctx); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if err := cache.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, _, err := cache.Get(ctx, "a"); !errors.Is(err, ErrBackendClosed) {
		t.Fatalf("Get(closed) error = %v", err)
	}
}

func TestCacheBackendSpecificBehaviorAndKeyPrefix(t *testing.T) {
	ctx := context.Background()
	redis := NewRedisCache(RedisOptions{BackendOptions: BackendOptions{KeyPrefix: "site", Version: 2}})
	if key := redis.Key("answer"); key != "site:2:answer" {
		t.Fatalf("redis key = %q", key)
	}
	if err := redis.Set(ctx, "answer", int64(42), time.Minute); err != nil {
		t.Fatalf("redis Set() error = %v", err)
	}
	if value, err := redis.Increment(ctx, "answer", 1); err != nil || value != 43 {
		t.Fatalf("redis Increment() = %d, %v", value, err)
	}
	file := NewFileCache(FileOptions{Directory: t.TempDir()})
	if file.Directory == "" {
		t.Fatal("file cache directory missing")
	}
	database := NewDatabaseCache(DatabaseOptions{})
	if database.Table != "gogo_cache" {
		t.Fatalf("database cache table = %q", database.Table)
	}
	memcached := NewMemcachedCache(MemcachedOptions{Servers: []string{"127.0.0.1:11211"}})
	if len(memcached.Servers) != 1 {
		t.Fatalf("memcached servers = %#v", memcached.Servers)
	}
	dummy := NewDummyCache()
	if err := dummy.Set(ctx, "ignored", "value", time.Minute); err != nil {
		t.Fatalf("dummy Set() error = %v", err)
	}
	if _, ok, err := dummy.Get(ctx, "ignored"); err != nil || ok {
		t.Fatalf("dummy Get() ok=%v err=%v", ok, err)
	}
}

func TestRedisCacheIntegrationGate(t *testing.T) {
	if os.Getenv("GOGO_TEST_REDIS_ADDR") == "" {
		t.Skip("set GOGO_TEST_REDIS_ADDR to run Redis cache integration tests")
	}
	cache := NewRedisCache(RedisOptions{Address: os.Getenv("GOGO_TEST_REDIS_ADDR")})
	if err := cache.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
}
