package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Rate stores a request limit and fixed window.
type Rate struct {
	Limit  int
	Window time.Duration
}

// ParseRate parses values such as "100/minute", "10/s", and "1000/day".
func ParseRate(value string) (Rate, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 2 {
		return Rate{}, fmt.Errorf("%w: invalid rate", ErrThrottled)
	}
	limit, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || limit <= 0 {
		return Rate{}, fmt.Errorf("%w: invalid rate limit", ErrThrottled)
	}
	window, ok := rateWindow(parts[1])
	if !ok {
		return Rate{}, fmt.Errorf("%w: invalid rate window", ErrThrottled)
	}
	return Rate{Limit: limit, Window: window}, nil
}

// ThrottleError carries retry timing for throttled requests.
type ThrottleError struct {
	RetryAfter time.Duration
}

func (e *ThrottleError) Error() string {
	return ErrThrottled.Error()
}

func (e *ThrottleError) Unwrap() error {
	return ErrThrottled
}

// ThrottleDecision is one throttle check result.
type ThrottleDecision struct {
	Allowed    bool
	RetryAfter time.Duration
	Key        string
	Scope      string
}

// Throttle checks whether a request may continue.
type Throttle interface {
	Allow(context.Context, *Request) (ThrottleDecision, error)
}

// ThrottleStore stores request counters.
type ThrottleStore interface {
	Hit(context.Context, string, time.Time, time.Duration) (int, time.Time, error)
}

type throttleBucket struct {
	StartedAt time.Time
	Count     int
}

// MemoryThrottleStore stores throttle counters in memory.
type MemoryThrottleStore struct {
	mu      sync.Mutex
	buckets map[string]throttleBucket
}

// NewMemoryThrottleStore creates an in-memory throttle store.
func NewMemoryThrottleStore() *MemoryThrottleStore {
	return &MemoryThrottleStore{buckets: map[string]throttleBucket{}}
}

// Hit increments one throttle key and returns its count and reset time.
func (s *MemoryThrottleStore) Hit(_ context.Context, key string, now time.Time, window time.Duration) (int, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.buckets[key]
	if bucket.StartedAt.IsZero() || !now.Before(bucket.StartedAt.Add(window)) {
		bucket = throttleBucket{StartedAt: now}
	}
	bucket.Count++
	s.buckets[key] = bucket
	return bucket.Count, bucket.StartedAt.Add(window), nil
}

// RateThrottle is a fixed-window throttle.
type RateThrottle struct {
	Scope    string
	Rate     Rate
	Store    ThrottleStore
	Identity func(*Request) string
	Now      func() time.Time
}

// AnonymousRateThrottle throttles anonymous requests by remote IP.
func AnonymousRateThrottle(rate Rate, store ThrottleStore) *RateThrottle {
	return &RateThrottle{Scope: "anon", Rate: rate, Store: store, Identity: anonymousThrottleIdentity}
}

// UserRateThrottle throttles authenticated users by user ID and anonymous users by IP.
func UserRateThrottle(rate Rate, store ThrottleStore) *RateThrottle {
	return &RateThrottle{Scope: "user", Rate: rate, Store: store, Identity: userThrottleIdentity}
}

// ScopedRateThrottle throttles requests by scope and request identity.
func ScopedRateThrottle(scope string, rate Rate, store ThrottleStore) *RateThrottle {
	return &RateThrottle{Scope: scope, Rate: rate, Store: store, Identity: userThrottleIdentity}
}

// Allow checks one request against the throttle store.
func (t *RateThrottle) Allow(ctx context.Context, request *Request) (ThrottleDecision, error) {
	if t == nil || t.Rate.Limit <= 0 || t.Rate.Window <= 0 {
		return ThrottleDecision{Allowed: true}, nil
	}
	store := t.Store
	if store == nil {
		store = NewMemoryThrottleStore()
		t.Store = store
	}
	identity := "anonymous"
	if t.Identity != nil {
		identity = t.Identity(request)
	}
	scope := t.Scope
	if scope == "" {
		scope = "default"
	}
	now := time.Now().UTC()
	if t.Now != nil {
		now = t.Now().UTC()
	}
	key := scope + ":" + identity
	count, reset, err := store.Hit(ctx, key, now, t.Rate.Window)
	if err != nil {
		return ThrottleDecision{}, err
	}
	if count > t.Rate.Limit {
		retry := reset.Sub(now)
		if retry < 0 {
			retry = 0
		}
		return ThrottleDecision{Allowed: false, RetryAfter: retry, Key: key, Scope: scope}, nil
	}
	return ThrottleDecision{Allowed: true, Key: key, Scope: scope}, nil
}

// CheckThrottles creates an APIView throttle lifecycle hook.
func CheckThrottles(throttles ...Throttle) RequestHook {
	return func(ctx context.Context, request *Request) error {
		for _, throttle := range throttles {
			if throttle == nil {
				continue
			}
			decision, err := throttle.Allow(ctx, request)
			if err != nil {
				return err
			}
			if !decision.Allowed {
				return &ThrottleError{RetryAfter: decision.RetryAfter}
			}
		}
		return nil
	}
}

func anonymousThrottleIdentity(request *Request) string {
	if request == nil {
		return "unknown"
	}
	if ip := request.RemoteIP(); ip != "" {
		return "ip:" + ip
	}
	return "ip:unknown"
}

func userThrottleIdentity(request *Request) string {
	user := apiUser(request)
	if isAuthenticatedUser(user) {
		return "user:" + strconv.FormatInt(user.ID, 10)
	}
	return anonymousThrottleIdentity(request)
}

func rateWindow(value string) (time.Duration, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "s", "sec", "second", "seconds":
		return time.Second, true
	case "m", "min", "minute", "minutes":
		return time.Minute, true
	case "h", "hour", "hours":
		return time.Hour, true
	case "d", "day", "days":
		return 24 * time.Hour, true
	default:
		return 0, false
	}
}
