package cache

import (
	"context"
	"net/http"
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
