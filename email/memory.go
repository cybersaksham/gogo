package email

import (
	"context"
	"sync"
)

type MemoryBackend struct {
	mu       sync.Mutex
	messages []Message
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{}
}

func (b *MemoryBackend) SendMessages(ctx context.Context, messages []Message) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, message := range messages {
		b.messages = append(b.messages, message.Clone())
	}
	return len(messages), nil
}

func (b *MemoryBackend) Messages() []Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	messages := make([]Message, len(b.messages))
	for i, message := range b.messages {
		messages[i] = message.Clone()
	}
	return messages
}

func (b *MemoryBackend) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = nil
}

func (b *MemoryBackend) Close() error { return nil }
