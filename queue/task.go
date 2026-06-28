package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDuplicateTask = errors.New("duplicate task")
	ErrInvalidTask   = errors.New("invalid task")
)

// TaskFunc is the executable task function contract.
type TaskFunc func(context.Context, ...any) (any, error)

// AckPolicy controls when workers acknowledge broker deliveries.
type AckPolicy string

const (
	AckEarly  AckPolicy = "early"
	AckLate   AckPolicy = "late"
	AckManual AckPolicy = "manual"
)

// RateLimit describes a per-task execution limit.
type RateLimit struct {
	Limit  int
	Period time.Duration
}

// TaskOptions contains Celery-style task execution options.
type TaskOptions struct {
	Serializer        string
	Queue             string
	RoutingKey        string
	Priority          int
	MaxRetries        int
	DefaultRetryDelay time.Duration
	RetryBackoff      bool
	RetryJitter       bool
	SoftTimeout       time.Duration
	HardTimeout       time.Duration
	RateLimit         RateLimit
	AckPolicy         AckPolicy
	IgnoreResult      bool
	TrackStarted      bool
}

// Task is one registered queue task.
type Task struct {
	Name    string
	Func    TaskFunc
	Options TaskOptions
}

// TaskDefinition is discovered from installed apps.
type TaskDefinition struct {
	Name    string
	Func    TaskFunc
	Options TaskOptions
}

// TaskProvider is implemented by installed apps that expose queue tasks.
type TaskProvider interface {
	QueueTasks() []TaskDefinition
}

func validateTask(name string, fn TaskFunc) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidTask)
	}
	if fn == nil {
		return fmt.Errorf("%w: function is required", ErrInvalidTask)
	}
	return nil
}
