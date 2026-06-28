package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cybersaksham/gogo/queue"
)

func TestRedisBackendKeysAndResultBehavior(t *testing.T) {
	backend := NewBackend(Config{Prefix: "gogo"})
	keys := backend.Keys("task-1")
	if keys.Result != "gogo:result:task-1" || keys.Children != "gogo:result:task-1:children" || keys.Group != "gogo:group:task-1" || keys.Chord != "gogo:chord:task-1" {
		t.Fatalf("keys = %#v", keys)
	}
	result := queue.Result{TaskID: "task-1", State: queue.StateSuccess, Result: "ok", Traceback: "trace", Children: []string{"child"}}
	if err := backend.StoreResult(context.Background(), result); err != nil {
		t.Fatalf("StoreResult() error = %v", err)
	}
	stored, err := backend.GetResult(context.Background(), "task-1")
	if err != nil || stored.State != queue.StateSuccess || stored.Result != "ok" || stored.Traceback != "trace" {
		t.Fatalf("GetResult() = %#v, %v", stored, err)
	}
	count, err := backend.ChordCounter(context.Background(), "chord-1", 2)
	if err != nil || count != 2 {
		t.Fatalf("ChordCounter() = %d, %v", count, err)
	}
}

func TestRedisBackendIntegration(t *testing.T) {
	if os.Getenv("GOGO_TEST_REDIS_ADDR") == "" {
		t.Skip("set GOGO_TEST_REDIS_ADDR to run Redis backend integration tests")
	}
	t.Skip("real Redis backend integration is enabled when a Redis client dependency is configured")
}

func TestRedisBackendResultExpiry(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC)
	backend := NewBackend(Config{Now: func() time.Time { return now }})
	expires := now.Add(time.Minute)
	if err := backend.StoreResult(context.Background(), queue.Result{TaskID: "task-1", State: queue.StateSuccess, ExpiresAt: &expires}); err != nil {
		t.Fatalf("StoreResult() error = %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := backend.GetResult(context.Background(), "task-1"); err == nil {
		t.Fatalf("expired result should fail")
	}
}
