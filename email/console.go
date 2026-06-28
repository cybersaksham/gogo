package email

import (
	"context"
	"fmt"
	"io"
)

type ConsoleBackend struct {
	writer       io.Writer
	failSilently bool
}

func NewConsoleBackend(writer io.Writer) *ConsoleBackend {
	return &ConsoleBackend{writer: writer}
}

func (b *ConsoleBackend) SendMessages(ctx context.Context, messages []Message) (int, error) {
	if err := ctx.Err(); err != nil {
		return handleSendError(0, err, b.failSilently)
	}
	sent := 0
	for _, message := range messages {
		body, err := message.RenderMIME()
		if err != nil {
			return handleSendError(sent, err, b.failSilently)
		}
		if _, err := fmt.Fprintf(b.writer, "%s\n", body); err != nil {
			return handleSendError(sent, err, b.failSilently)
		}
		sent++
	}
	return sent, nil
}

func (b *ConsoleBackend) Close() error { return nil }
