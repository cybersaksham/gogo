package queue

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestMessageSignerValidTamperedAndExpiredEnvelope(t *testing.T) {
	now := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	signer := NewMessageSigner(MessageSignerOptions{
		PrimaryKeyID: "primary",
		Keys:         map[string][]byte{"primary": []byte("secret")},
		Now:          func() time.Time { return now },
		ReplayWindow: time.Minute,
	})
	envelope := Envelope{ID: "task-1", Name: "jobs.secure", Args: []any{"safe"}, CreatedAt: now}
	headers, err := signer.Sign(envelope)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if err := signer.Verify(envelope, headers); err != nil {
		t.Fatalf("Verify(valid) error = %v", err)
	}
	tampered := envelope
	tampered.Name = "jobs.tampered"
	if err := signer.Verify(tampered, headers); !errors.Is(err, ErrInvalidMessageSignature) {
		t.Fatalf("Verify(tampered) error = %v", err)
	}
	now = now.Add(2 * time.Minute)
	if err := signer.Verify(envelope, headers); !errors.Is(err, ErrMessageExpired) {
		t.Fatalf("Verify(expired) error = %v", err)
	}
}

func TestContentTypeAllowlistRejectsUnknownContent(t *testing.T) {
	allowlist := NewContentTypeAllowlist("application/json")
	if err := allowlist.Validate("application/json"); err != nil {
		t.Fatalf("Validate(json) error = %v", err)
	}
	if err := allowlist.Validate("application/gob"); !errors.Is(err, ErrRejectedContentType) {
		t.Fatalf("Validate(gob) error = %v", err)
	}
}

func TestBrokerTLSValidation(t *testing.T) {
	if err := ValidateBrokerTLS(BrokerTLSConfig{URL: "amqps://broker.internal", TLSEnabled: true, ServerName: "broker.internal"}); err != nil {
		t.Fatalf("ValidateBrokerTLS(valid) error = %v", err)
	}
	if err := ValidateBrokerTLS(BrokerTLSConfig{URL: "amqps://broker.internal", TLSEnabled: false}); !errors.Is(err, ErrInvalidBrokerTLS) {
		t.Fatalf("ValidateBrokerTLS(disabled amqps) error = %v", err)
	}
	if err := ValidateBrokerTLS(BrokerTLSConfig{URL: "amqps://broker.internal", TLSEnabled: true, InsecureSkipVerify: true}); !errors.Is(err, ErrInvalidBrokerTLS) {
		t.Fatalf("ValidateBrokerTLS(insecure) error = %v", err)
	}
}

func TestRedactionForSensitiveArgsKwargsAndEvents(t *testing.T) {
	redactor := NewRedactor(RedactorOptions{SensitiveKeys: []string{"password", "token"}})
	envelope := Envelope{
		ID:     "task-1",
		Name:   "jobs.login",
		Args:   []any{"user", Sensitive("plain-secret")},
		Kwargs: map[string]any{"password": "pw", "safe": "ok"},
	}
	redacted := redactor.RedactEnvelope(envelope)
	if redacted.Args[1] != RedactedValue || redacted.Kwargs["password"] != RedactedValue || redacted.Kwargs["safe"] != "ok" {
		t.Fatalf("redacted envelope = %#v", redacted)
	}
	event := redactor.RedactEvent(Event{Fields: map[string]any{"token": "abc", "count": 1}})
	if !reflect.DeepEqual(event.Fields, map[string]any{"token": RedactedValue, "count": 1}) {
		t.Fatalf("redacted event = %#v", event)
	}
	result := redactor.RedactResult(Result{Result: map[string]any{"token": "abc", "value": "ok"}})
	if !reflect.DeepEqual(result.Result, map[string]any{"token": RedactedValue, "value": "ok"}) {
		t.Fatalf("redacted result = %#v", result)
	}
}
