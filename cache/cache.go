package cache

import (
	"context"
	"net/http"
	"sync"
	"time"
)

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
