package queue

import (
	"errors"
	"testing"
	"time"
)

func TestRetryDelayBackoffJitterAndMaxRetries(t *testing.T) {
	options := TaskOptions{MaxRetries: 3, DefaultRetryDelay: time.Second, RetryBackoff: true}
	if !CanRetry(options, 0) || !CanRetry(options, 2) || CanRetry(options, 3) {
		t.Fatalf("CanRetry() did not respect max retries")
	}
	if delay := ComputeRetryDelay(options, 0, nil); delay != time.Second {
		t.Fatalf("retry delay attempt 0 = %s", delay)
	}
	if delay := ComputeRetryDelay(options, 2, nil); delay != 4*time.Second {
		t.Fatalf("retry delay attempt 2 = %s", delay)
	}
	options.RetryJitter = true
	delay := ComputeRetryDelay(options, 1, func(max time.Duration) time.Duration {
		return max / 2
	})
	if delay != time.Second {
		t.Fatalf("jitter delay = %s", delay)
	}
}

func TestRetryErrorOptions(t *testing.T) {
	base := errors.New("temporary")
	err := Retry(base, RetryCountdown(5*time.Second), RetryMaxRetries(7))
	retry, ok := AsRetry(err)
	if !ok {
		t.Fatalf("AsRetry() ok = false")
	}
	if !errors.Is(err, base) || retry.Countdown != 5*time.Second || retry.MaxRetries == nil || *retry.MaxRetries != 7 {
		t.Fatalf("retry error = %#v", retry)
	}
}

func TestRevocationRegistryByTaskAndStampedHeader(t *testing.T) {
	registry := NewRevocationRegistry()
	registry.RevokeTask("task-1")
	registry.RevokeStampedHeader("tenant", "blocked")
	if !registry.IsRevoked(Envelope{ID: "task-1"}) {
		t.Fatal("task ID revocation was not matched")
	}
	if !registry.IsRevoked(Envelope{ID: "task-2", Headers: map[string]string{"tenant": "blocked"}}) {
		t.Fatal("stamped header revocation was not matched")
	}
	if registry.IsRevoked(Envelope{ID: "task-3", Headers: map[string]string{"tenant": "allowed"}}) {
		t.Fatal("unexpected revocation match")
	}
	registry.ClearTask("task-1")
	if registry.IsRevoked(Envelope{ID: "task-1"}) {
		t.Fatal("task ID revocation was not cleared")
	}
}
