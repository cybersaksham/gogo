package email

import "context"

type DummyBackend struct {
	count int
}

func NewDummyBackend() *DummyBackend {
	return &DummyBackend{}
}

func (b *DummyBackend) SendMessages(ctx context.Context, messages []Message) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.count += len(messages)
	return len(messages), nil
}

func (b *DummyBackend) Count() int {
	return b.count
}

func (b *DummyBackend) Close() error { return nil }
