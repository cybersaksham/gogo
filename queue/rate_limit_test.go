package queue

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinWindowAndReturnsRetryDelay(t *testing.T) {
	now := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	limiter := NewRateLimiter(RateLimiterOptions{Now: func() time.Time { return now }})
	limit := RateLimit{Limit: 2, Period: time.Second}
	if allowed, delay := limiter.Allow("emails.send", limit); !allowed || delay != 0 {
		t.Fatalf("first Allow() = %v, %s", allowed, delay)
	}
	if allowed, delay := limiter.Allow("emails.send", limit); !allowed || delay != 0 {
		t.Fatalf("second Allow() = %v, %s", allowed, delay)
	}
	if allowed, delay := limiter.Allow("emails.send", limit); allowed || delay != time.Second {
		t.Fatalf("limited Allow() = %v, %s", allowed, delay)
	}
	now = now.Add(time.Second)
	if allowed, delay := limiter.Allow("emails.send", limit); !allowed || delay != 0 {
		t.Fatalf("after window Allow() = %v, %s", allowed, delay)
	}
}

func TestRateLimiterTreatsEmptyLimitAsUnlimited(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterOptions{})
	if allowed, delay := limiter.Allow("jobs.any", RateLimit{}); !allowed || delay != 0 {
		t.Fatalf("unlimited Allow() = %v, %s", allowed, delay)
	}
}
