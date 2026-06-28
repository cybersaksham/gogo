package email

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type FileBackend struct {
	dir          string
	counter      atomic.Int64
	failSilently bool
}

func NewFileBackend(dir string) *FileBackend {
	return &FileBackend{dir: dir}
}

func (b *FileBackend) SendMessages(ctx context.Context, messages []Message) (int, error) {
	if err := os.MkdirAll(b.dir, 0o755); err != nil {
		return handleSendError(0, err, b.failSilently)
	}
	sent := 0
	for _, message := range messages {
		if err := ctx.Err(); err != nil {
			return handleSendError(sent, err, b.failSilently)
		}
		body, err := message.RenderMIME()
		if err != nil {
			return handleSendError(sent, err, b.failSilently)
		}
		name := fmt.Sprintf("%d-%06d.eml", time.Now().UTC().UnixNano(), b.counter.Add(1))
		if err := os.WriteFile(filepath.Join(b.dir, name), body, 0o600); err != nil {
			return handleSendError(sent, err, b.failSilently)
		}
		sent++
	}
	return sent, nil
}

func (b *FileBackend) Close() error { return nil }
