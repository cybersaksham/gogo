package redis

import (
	"context"
	"strings"
	"time"

	"github.com/cybersaksham/gogo/queue"
	"github.com/cybersaksham/gogo/queue/brokers"
)

type Config struct {
	Prefix            string
	VisibilityTimeout time.Duration
	PriorityBuckets   int
}

type Broker struct {
	config Config
	memory *brokers.MemoryBroker
}

type Keys struct {
	Ready      string
	Delayed    string
	Unacked    string
	DeadLetter string

	prefix string
	queue  string
}

func NewBroker(config Config) *Broker {
	if config.Prefix == "" {
		config.Prefix = "gogo"
	}
	if config.PriorityBuckets <= 0 {
		config.PriorityBuckets = 10
	}
	return &Broker{
		config: config,
		memory: brokers.NewMemoryBroker(brokers.MemoryOptions{VisibilityTimeout: config.VisibilityTimeout}),
	}
}

func (b *Broker) Keys(queueName string) Keys {
	queueName = strings.TrimSpace(queueName)
	if queueName == "" {
		queueName = "default"
	}
	base := b.config.Prefix + ":queue:" + queueName
	return Keys{
		Ready:      base + ":ready",
		Delayed:    base + ":delayed",
		Unacked:    base + ":unacked",
		DeadLetter: base + ":dead",
		prefix:     b.config.Prefix,
		queue:      queueName,
	}
}

func (k Keys) Priority(priority int) string {
	if priority < 0 {
		priority = 0
	}
	return k.prefix + ":queue:" + k.queue + ":priority:" + strconvItoa(priority)
}

func EncodeEnvelope(envelope queue.Envelope) ([]byte, error) {
	registry := queue.NewSerializationRegistry(queue.SerializationOptions{})
	payload, err := registry.Encode("json", envelope, queue.CompressionNone)
	if err != nil {
		return nil, err
	}
	return payload.Body, nil
}

func DecodeEnvelope(data []byte) (queue.Envelope, error) {
	registry := queue.NewSerializationRegistry(queue.SerializationOptions{})
	payload := queue.Payload{Serializer: "json", Compression: queue.CompressionNone, Body: data}
	var envelope queue.Envelope
	if err := registry.Decode(payload, &envelope); err != nil {
		return queue.Envelope{}, err
	}
	return envelope, nil
}

func (b *Broker) Publish(ctx context.Context, queueName string, envelope queue.Envelope, options brokers.PublishOptions) (brokers.Message, error) {
	return b.memory.Publish(ctx, queueName, envelope, options)
}

func (b *Broker) Consume(ctx context.Context, queueName string, options brokers.ConsumeOptions) (brokers.Message, error) {
	return b.memory.Consume(ctx, queueName, options)
}

func (b *Broker) Ack(ctx context.Context, message brokers.Message) error {
	return b.memory.Ack(ctx, message)
}

func (b *Broker) Nack(ctx context.Context, message brokers.Message, requeue bool) error {
	return b.memory.Nack(ctx, message, requeue)
}

func (b *Broker) Requeue(ctx context.Context, message brokers.Message, delay time.Duration) error {
	return b.memory.Requeue(ctx, message, delay)
}

func (b *Broker) DeclareQueue(ctx context.Context, queueName string, options brokers.QueueOptions) error {
	return b.memory.DeclareQueue(ctx, queueName, options)
}

func (b *Broker) PurgeQueue(ctx context.Context, queueName string) (int, error) {
	return b.memory.PurgeQueue(ctx, queueName)
}

func (b *Broker) InspectQueues(ctx context.Context) ([]brokers.QueueInfo, error) {
	return b.memory.InspectQueues(ctx)
}

func (b *Broker) Close() error {
	return b.memory.Close()
}

func strconvItoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
