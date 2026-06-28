package redis

import (
	"context"
	"time"

	"github.com/cybersaksham/gogo/queue"
	"github.com/cybersaksham/gogo/queue/backends"
)

type Config struct {
	Prefix string
	Now    func() time.Time
}

type Backend struct {
	config Config
	memory *backends.MemoryBackend
}

type Keys struct {
	Result   string
	Children string
	Group    string
	Chord    string
}

func NewBackend(config Config) *Backend {
	if config.Prefix == "" {
		config.Prefix = "gogo"
	}
	return &Backend{
		config: config,
		memory: backends.NewMemoryBackend(backends.MemoryOptions{Now: config.Now}),
	}
}

func (b *Backend) Keys(id string) Keys {
	return Keys{
		Result:   b.config.Prefix + ":result:" + id,
		Children: b.config.Prefix + ":result:" + id + ":children",
		Group:    b.config.Prefix + ":group:" + id,
		Chord:    b.config.Prefix + ":chord:" + id,
	}
}

func (b *Backend) StoreResult(ctx context.Context, result queue.Result) error {
	return b.memory.StoreResult(ctx, result)
}

func (b *Backend) GetResult(ctx context.Context, taskID string) (queue.Result, error) {
	return b.memory.GetResult(ctx, taskID)
}

func (b *Backend) Forget(ctx context.Context, taskID string) error {
	return b.memory.Forget(ctx, taskID)
}

func (b *Backend) Wait(ctx context.Context, taskID string, timeout time.Duration) (queue.Result, error) {
	return b.memory.Wait(ctx, taskID, timeout)
}

func (b *Backend) Children(ctx context.Context, taskID string) ([]string, error) {
	return b.memory.Children(ctx, taskID)
}

func (b *Backend) GroupResult(ctx context.Context, groupID string, children []string) (queue.GroupResult, error) {
	return b.memory.GroupResult(ctx, groupID, children)
}

func (b *Backend) ChordCounter(ctx context.Context, chordID string, delta int) (int, error) {
	return b.memory.ChordCounter(ctx, chordID, delta)
}
