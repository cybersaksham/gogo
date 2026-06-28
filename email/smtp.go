package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPSender interface {
	Open(context.Context) error
	Send(context.Context, Message, []byte) error
	Close() error
}

type SMTPOptions struct {
	Host         string
	Port         int
	Username     string
	Password     string
	UseTLS       bool
	FailSilently bool
	Sender       SMTPSender
}

type SMTPBackend struct {
	options SMTPOptions
	sender  SMTPSender
	open    bool
}

func NewSMTPBackend(options SMTPOptions) *SMTPBackend {
	sender := options.Sender
	if sender == nil {
		sender = smtpSender{options: options}
	}
	return &SMTPBackend{options: options, sender: sender}
}

func (b *SMTPBackend) Open(ctx context.Context) error {
	if b.open {
		return nil
	}
	if err := b.sender.Open(ctx); err != nil {
		return err
	}
	b.open = true
	return nil
}

func (b *SMTPBackend) SendMessages(ctx context.Context, messages []Message) (int, error) {
	if err := b.Open(ctx); err != nil {
		return handleSendError(0, err, b.options.FailSilently)
	}
	sent := 0
	for _, message := range messages {
		body, err := message.RenderMIME()
		if err != nil {
			return handleSendError(sent, err, b.options.FailSilently)
		}
		if err := b.sender.Send(ctx, message, body); err != nil {
			return handleSendError(sent, err, b.options.FailSilently)
		}
		sent++
	}
	return sent, nil
}

func (b *SMTPBackend) Close() error {
	if !b.open {
		return nil
	}
	b.open = false
	return b.sender.Close()
}

type smtpSender struct {
	options SMTPOptions
}

func (s smtpSender) Open(context.Context) error {
	if s.options.Host == "" {
		return fmt.Errorf("smtp host is required")
	}
	return nil
}

func (s smtpSender) Send(_ context.Context, message Message, body []byte) error {
	address := fmt.Sprintf("%s:%d", s.options.Host, s.options.Port)
	auth := smtp.Auth(nil)
	if s.options.Username != "" {
		auth = smtp.PlainAuth("", s.options.Username, s.options.Password, s.options.Host)
	}
	return smtp.SendMail(address, auth, message.From, message.Recipients(), body)
}

func (s smtpSender) Close() error {
	return nil
}
