package queue

import (
	"context"
	"errors"
	"time"
)

var (
	ErrQueueEmpty   = errors.New("queue empty")
	ErrBrokerClosed = errors.New("broker closed")
)

// Broker is the stable worker-facing queue broker contract.
type Broker interface {
	Publish(context.Context, string, Envelope, BrokerPublishOptions) (BrokerMessage, error)
	Consume(context.Context, string, BrokerConsumeOptions) (BrokerMessage, error)
	Ack(context.Context, BrokerMessage) error
	Nack(context.Context, BrokerMessage, bool) error
	Requeue(context.Context, BrokerMessage, time.Duration) error
	DeclareQueue(context.Context, string, BrokerQueueOptions) error
	PurgeQueue(context.Context, string) (int, error)
	InspectQueues(context.Context) ([]BrokerQueueInfo, error)
	Close() error
}

type BrokerPublishOptions struct {
	Priority   int
	RoutingKey string
	Headers    map[string]string
}

type BrokerConsumeOptions struct {
	VisibilityTimeout time.Duration
}

type BrokerQueueOptions struct {
	Durable           bool
	VisibilityTimeout time.Duration
}

type BrokerMessage struct {
	DeliveryID string
	Queue      string
	Envelope   Envelope
	Priority   int
	Attempts   int
	VisibleAt  time.Time
	Deadline   time.Time
}

type BrokerQueueInfo struct {
	Name     string
	Ready    int
	InFlight int
	Durable  bool
}

// ResultBackend is the worker-facing result storage contract.
type ResultBackend interface {
	StoreResult(context.Context, Result) error
	GetResult(context.Context, string) (Result, error)
	Forget(context.Context, string) error
	Wait(context.Context, string, time.Duration) (Result, error)
	Children(context.Context, string) ([]string, error)
	GroupResult(context.Context, string, []string) (GroupResult, error)
	ChordCounter(context.Context, string, int) (int, error)
}
