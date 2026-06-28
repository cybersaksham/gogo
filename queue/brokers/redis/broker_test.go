package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cybersaksham/gogo/queue"
	"github.com/cybersaksham/gogo/queue/brokers"
)

func TestRedisKeyNamingAndEnvelopeEncoding(t *testing.T) {
	broker := NewBroker(Config{Prefix: "gogo", VisibilityTimeout: time.Minute, PriorityBuckets: 10})
	keys := broker.Keys("emails")
	if keys.Ready != "gogo:queue:emails:ready" || keys.Delayed != "gogo:queue:emails:delayed" || keys.Unacked != "gogo:queue:emails:unacked" || keys.DeadLetter != "gogo:queue:emails:dead" {
		t.Fatalf("keys = %#v", keys)
	}
	if got := keys.Priority(9); got != "gogo:queue:emails:priority:9" {
		t.Fatalf("priority key = %q", got)
	}

	envelope := queue.Envelope{ID: "task-1", Name: "blog.publish", Queue: "emails", Priority: 9}
	encoded, err := EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope() error = %v", err)
	}
	decoded, err := DecodeEnvelope(encoded)
	if err != nil {
		t.Fatalf("DecodeEnvelope() error = %v", err)
	}
	if decoded.ID != envelope.ID || decoded.Name != envelope.Name || decoded.Queue != envelope.Queue || decoded.Priority != envelope.Priority {
		t.Fatalf("decoded = %#v", decoded)
	}
}

func TestRedisBrokerImplementsBrokerBehavior(t *testing.T) {
	ctx := context.Background()
	var _ brokers.Broker = NewBroker(Config{Prefix: "gogo"})
	broker := NewBroker(Config{Prefix: "gogo"})
	if err := broker.DeclareQueue(ctx, "default", brokers.QueueOptions{}); err != nil {
		t.Fatalf("DeclareQueue() error = %v", err)
	}
	if _, err := broker.Publish(ctx, "default", queue.Envelope{ID: "task-1", Name: "blog.publish"}, brokers.PublishOptions{Priority: 5}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	message, err := broker.Consume(ctx, "default", brokers.ConsumeOptions{})
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if message.Priority != 5 || message.Envelope.ID != "task-1" {
		t.Fatalf("message = %#v", message)
	}
	if err := broker.Ack(ctx, message); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}
}

func TestRedisBrokerIntegration(t *testing.T) {
	if os.Getenv("GOGO_TEST_REDIS_ADDR") == "" {
		t.Skip("set GOGO_TEST_REDIS_ADDR to run Redis broker integration tests")
	}
	t.Skip("real Redis client integration is enabled when a Redis client dependency is configured")
}
