package cache

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	ErrKeyNotFound       = errors.New("cache key not found")
	ErrValueNotInteger   = errors.New("cache value is not an integer")
	ErrBackendClosed     = errors.New("cache backend closed")
	ErrInvalidCacheValue = errors.New("invalid cache value")
)

// Backend is the Django-style general cache contract.
type Backend interface {
	Get(context.Context, string) (any, bool, error)
	Set(context.Context, string, any, time.Duration) error
	Add(context.Context, string, any, time.Duration) (bool, error)
	GetOrSet(context.Context, string, func(context.Context) (any, error), time.Duration) (any, error)
	Delete(context.Context, string) (bool, error)
	Clear(context.Context) error
	Touch(context.Context, string, time.Duration) (bool, error)
	Increment(context.Context, string, int64) (int64, error)
	Decrement(context.Context, string, int64) (int64, error)
	GetMany(context.Context, []string) (map[string]any, error)
	SetMany(context.Context, map[string]any, time.Duration) error
	DeleteMany(context.Context, []string) (int, error)
	Close() error
}

type BackendOptions struct {
	KeyPrefix string
	Version   int
}

func BuildKey(options BackendOptions, key string) string {
	version := options.Version
	if version == 0 {
		version = 1
	}
	if options.KeyPrefix == "" {
		return fmt.Sprintf("%d:%s", version, key)
	}
	return fmt.Sprintf("%s:%d:%s", options.KeyPrefix, version, key)
}

func asInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	case int32:
		return int64(typed), nil
	case uint:
		return int64(typed), nil
	case uint64:
		if typed > uint64(^uint(0)>>1) {
			return 0, ErrValueNotInteger
		}
		return int64(typed), nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, ErrValueNotInteger
	}
}

// Entry is a cached HTTP response payload.
type Entry struct {
	Status int
	Header http.Header
	Body   []byte
}

// Store is the minimal cache backend contract used by HTTP middleware.
type Store interface {
	Get(context.Context, string) (Entry, bool, error)
	Set(context.Context, string, Entry, time.Duration) error
}

// MemoryStore is a local in-memory cache backend.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]memoryItem
}

type memoryItem struct {
	entry   Entry
	expires time.Time
}

// NewMemoryStore creates an empty in-memory cache.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]memoryItem)}
}

// Get returns one cache entry.
func (s *MemoryStore) Get(_ context.Context, key string) (Entry, bool, error) {
	now := time.Now()

	s.mu.RLock()
	item, ok := s.items[key]
	s.mu.RUnlock()
	if !ok {
		return Entry{}, false, nil
	}
	if !item.expires.IsZero() && now.After(item.expires) {
		s.mu.Lock()
		delete(s.items, key)
		s.mu.Unlock()
		return Entry{}, false, nil
	}
	return cloneEntry(item.entry), true, nil
}

// Set stores one cache entry.
func (s *MemoryStore) Set(_ context.Context, key string, entry Entry, ttl time.Duration) error {
	item := memoryItem{entry: cloneEntry(entry)}
	if ttl != 0 {
		item.expires = time.Now().Add(ttl)
	}

	s.mu.Lock()
	s.items[key] = item
	s.mu.Unlock()
	return nil
}

func cloneEntry(entry Entry) Entry {
	return Entry{
		Status: entry.Status,
		Header: entry.Header.Clone(),
		Body:   append([]byte(nil), entry.Body...),
	}
}
