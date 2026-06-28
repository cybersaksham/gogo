package email

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

var errFakeSMTP = errors.New("smtp failure")

func TestEmailMessageMIMERenderingAlternativesAttachmentsAndHeaders(t *testing.T) {
	message := Message{
		Subject: "Welcome",
		Body:    "plain body",
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"hidden@example.com"},
		ReplyTo: []string{"reply@example.com"},
		Headers: map[string]string{"X-Test": "yes"},
	}
	message.AddAlternative("text/html", "<strong>html body</strong>")
	message.Attach("report.txt", "text/plain", []byte("attachment body"))
	rendered, err := message.RenderMIME()
	if err != nil {
		t.Fatalf("RenderMIME() error = %v", err)
	}
	body := string(rendered)
	for _, want := range []string{"Subject: Welcome", "To: to@example.com", "Cc: cc@example.com", "Reply-To: reply@example.com", "X-Test: yes", "plain body", "<strong>html body</strong>", "report.txt", "attachment body"} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered MIME missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "hidden@example.com") {
		t.Fatalf("Bcc leaked into MIME headers:\n%s", body)
	}
}

func TestConsoleFileMemoryAndDummyBackends(t *testing.T) {
	ctx := context.Background()
	message := Message{Subject: "Hello", Body: "body", From: "from@example.com", To: []string{"to@example.com"}}
	var console bytes.Buffer
	consoleBackend := NewConsoleBackend(&console)
	if count, err := consoleBackend.SendMessages(ctx, []Message{message}); err != nil || count != 1 {
		t.Fatalf("console SendMessages() = %d, %v", count, err)
	}
	if !strings.Contains(console.String(), "Subject: Hello") {
		t.Fatalf("console output = %q", console.String())
	}
	dir := t.TempDir()
	fileBackend := NewFileBackend(dir)
	if count, err := fileBackend.SendMessages(ctx, []Message{message}); err != nil || count != 1 {
		t.Fatalf("file SendMessages() = %d, %v", count, err)
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 1 {
		t.Fatalf("file backend wrote %d files", len(files))
	}
	memoryBackend := NewMemoryBackend()
	if count, err := memoryBackend.SendMessages(ctx, []Message{message}); err != nil || count != 1 {
		t.Fatalf("memory SendMessages() = %d, %v", count, err)
	}
	if len(memoryBackend.Messages()) != 1 {
		t.Fatalf("memory messages = %#v", memoryBackend.Messages())
	}
	dummy := NewDummyBackend()
	if count, err := dummy.SendMessages(ctx, []Message{message}); err != nil || count != 1 || dummy.Count() != 1 {
		t.Fatalf("dummy SendMessages() = %d, %v count=%d", count, err, dummy.Count())
	}
}

func TestSMTPBackendConnectionReuseAndFailSilently(t *testing.T) {
	ctx := context.Background()
	sender := &fakeSMTPSender{}
	backend := NewSMTPBackend(SMTPOptions{Sender: sender})
	messages := []Message{
		{Subject: "One", Body: "body", From: "from@example.com", To: []string{"one@example.com"}},
		{Subject: "Two", Body: "body", From: "from@example.com", To: []string{"two@example.com"}},
	}
	if count, err := backend.SendMessages(ctx, messages); err != nil || count != 2 {
		t.Fatalf("smtp SendMessages() = %d, %v", count, err)
	}
	if sender.opens != 1 || sender.sends != 2 {
		t.Fatalf("sender opens=%d sends=%d", sender.opens, sender.sends)
	}
	if err := backend.Close(); err != nil || sender.closes != 1 {
		t.Fatalf("Close() err=%v closes=%d", err, sender.closes)
	}
	failing := NewSMTPBackend(SMTPOptions{Sender: &fakeSMTPSender{fail: true}, FailSilently: true})
	if count, err := failing.SendMessages(ctx, []Message{messages[0]}); err != nil || count != 0 {
		t.Fatalf("fail silently SendMessages() = %d, %v", count, err)
	}
}

type fakeSMTPSender struct {
	opens  int
	sends  int
	closes int
	fail   bool
}

func (s *fakeSMTPSender) Open(context.Context) error {
	s.opens++
	return nil
}

func (s *fakeSMTPSender) Send(context.Context, Message, []byte) error {
	if s.fail {
		return errFakeSMTP
	}
	s.sends++
	return nil
}

func (s *fakeSMTPSender) Close() error {
	s.closes++
	return nil
}
