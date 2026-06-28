package queue

import (
	"sync"
	"time"
)

type RateLimiterOptions struct {
	Now func() time.Time
}

type RateLimiter struct {
	mu     sync.Mutex
	now    func() time.Time
	events map[string][]time.Time
}

func NewRateLimiter(options RateLimiterOptions) *RateLimiter {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &RateLimiter{now: now, events: map[string][]time.Time{}}
}

func (l *RateLimiter) Allow(taskName string, limit RateLimit) (bool, time.Duration) {
	if limit.Limit <= 0 {
		return true, 0
	}
	period := limit.Period
	if period <= 0 {
		period = time.Second
	}
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	events := l.events[taskName]
	windowStart := now.Add(-period)
	kept := events[:0]
	for _, event := range events {
		if event.After(windowStart) {
			kept = append(kept, event)
		}
	}
	if len(kept) < limit.Limit {
		kept = append(kept, now)
		l.events[taskName] = kept
		return true, 0
	}
	l.events[taskName] = kept
	retryAt := kept[0].Add(period)
	delay := retryAt.Sub(now)
	if delay < 0 {
		delay = 0
	}
	return false, delay
}

func (l *RateLimiter) Reset(taskName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.events, taskName)
}
