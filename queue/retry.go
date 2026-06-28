package queue

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrRetryRequested = errors.New("retry requested")
	ErrSoftTimeout    = errors.New("soft timeout exceeded")
	ErrHardTimeout    = errors.New("hard timeout exceeded")
)

type RetryError struct {
	Err        error
	Countdown  time.Duration
	ETA        *time.Time
	MaxRetries *int
}

func (e *RetryError) Error() string {
	if e == nil || e.Err == nil {
		return ErrRetryRequested.Error()
	}
	return e.Err.Error()
}

func (e *RetryError) Unwrap() error {
	if e == nil || e.Err == nil {
		return ErrRetryRequested
	}
	return e.Err
}

type RetryOption func(*RetryError)

func Retry(err error, options ...RetryOption) error {
	if err == nil {
		err = ErrRetryRequested
	}
	retry := &RetryError{Err: err}
	for _, option := range options {
		if option != nil {
			option(retry)
		}
	}
	return retry
}

func RetryCountdown(countdown time.Duration) RetryOption {
	return func(retry *RetryError) {
		retry.Countdown = countdown
	}
}

func RetryETA(eta time.Time) RetryOption {
	return func(retry *RetryError) {
		retry.ETA = &eta
	}
}

func RetryMaxRetries(maxRetries int) RetryOption {
	return func(retry *RetryError) {
		retry.MaxRetries = &maxRetries
	}
}

func AsRetry(err error) (*RetryError, bool) {
	var retry *RetryError
	if errors.As(err, &retry) {
		return retry, true
	}
	return nil, false
}

func CanRetry(options TaskOptions, currentRetries int) bool {
	return options.MaxRetries > currentRetries
}

func ComputeRetryDelay(options TaskOptions, currentRetries int, jitter func(time.Duration) time.Duration) time.Duration {
	delay := options.DefaultRetryDelay
	if delay <= 0 {
		delay = time.Second
	}
	if options.RetryBackoff {
		for i := 0; i < currentRetries; i++ {
			if delay > time.Duration(1<<62) {
				break
			}
			delay *= 2
		}
	}
	if options.RetryJitter {
		if jitter != nil {
			return jitter(delay)
		}
		if delay <= 0 {
			return 0
		}
		return time.Duration(rand.Int63n(int64(delay) + 1))
	}
	return delay
}

type RevocationRegistry struct {
	mu      sync.RWMutex
	tasks   map[string]struct{}
	headers map[string]map[string]struct{}
}

func NewRevocationRegistry() *RevocationRegistry {
	return &RevocationRegistry{
		tasks:   map[string]struct{}{},
		headers: map[string]map[string]struct{}{},
	}
}

func (r *RevocationRegistry) RevokeTask(taskID string) {
	if taskID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[taskID] = struct{}{}
}

func (r *RevocationRegistry) ClearTask(taskID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, taskID)
}

func (r *RevocationRegistry) RevokeStampedHeader(name string, value string) {
	if name == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	values := r.headers[name]
	if values == nil {
		values = map[string]struct{}{}
		r.headers[name] = values
	}
	values[value] = struct{}{}
}

func (r *RevocationRegistry) ClearStampedHeader(name string, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if values := r.headers[name]; values != nil {
		delete(values, value)
		if len(values) == 0 {
			delete(r.headers, name)
		}
	}
}

func (r *RevocationRegistry) IsRevoked(envelope Envelope) bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.tasks[envelope.ID]; ok {
		return true
	}
	for name, value := range envelope.Headers {
		if values := r.headers[name]; values != nil {
			if _, ok := values[value]; ok {
				return true
			}
		}
	}
	return false
}
