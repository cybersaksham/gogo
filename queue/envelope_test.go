package queue

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestEnvelopeSerializationCompressionTrustedRawGobAndCustom(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC)
	eta := now.Add(5 * time.Minute)
	expires := now.Add(time.Hour)
	envelope := Envelope{
		ID:            "task-1",
		RootID:        "root-1",
		ParentID:      "parent-1",
		GroupID:       "group-1",
		ChordID:       "chord-1",
		Name:          "blog.publish",
		Args:          []any{"go", float64(7)},
		Kwargs:        map[string]any{"force": true},
		Headers:       map[string]string{"trace": "abc"},
		Retries:       2,
		ETA:           &eta,
		Expires:       &expires,
		Queue:         "emails",
		Priority:      9,
		ReplyTo:       "reply",
		CorrelationID: "corr",
		CreatedAt:     now,
	}

	registry := NewSerializationRegistry(SerializationOptions{})
	payload, err := registry.Encode("json", envelope, CompressionGzip)
	if err != nil {
		t.Fatalf("Encode(json) error = %v", err)
	}
	if payload.Serializer != "json" || payload.Compression != CompressionGzip || len(payload.Body) == 0 {
		t.Fatalf("payload = %#v", payload)
	}
	var decoded Envelope
	if err := registry.Decode(payload, &decoded); err != nil {
		t.Fatalf("Decode(json) error = %v", err)
	}
	if decoded.ID != envelope.ID || decoded.Name != envelope.Name || decoded.Queue != "emails" || decoded.Headers["trace"] != "abc" || decoded.ETA.IsZero() {
		t.Fatalf("decoded envelope = %#v", decoded)
	}

	if _, err := registry.Encode("gob", envelope, CompressionNone); !errors.Is(err, ErrUntrustedSerializer) {
		t.Fatalf("Encode(gob) error = %v, want ErrUntrustedSerializer", err)
	}
	trusted := NewSerializationRegistry(SerializationOptions{AllowUntrustedSerializers: []string{"gob"}})
	gobPayload, err := trusted.Encode("gob", envelope, CompressionNone)
	if err != nil {
		t.Fatalf("Encode(trusted gob) error = %v", err)
	}
	var gobDecoded Envelope
	if err := trusted.Decode(gobPayload, &gobDecoded); err != nil {
		t.Fatalf("Decode(gob) error = %v", err)
	}
	if gobDecoded.ID != "task-1" {
		t.Fatalf("gob decoded = %#v", gobDecoded)
	}

	rawPayload, err := registry.Encode("raw", []byte("raw-body"), CompressionNone)
	if err != nil {
		t.Fatalf("Encode(raw) error = %v", err)
	}
	var raw []byte
	if err := registry.Decode(rawPayload, &raw); err != nil {
		t.Fatalf("Decode(raw) error = %v", err)
	}
	if !bytes.Equal(raw, []byte("raw-body")) {
		t.Fatalf("raw = %q", raw)
	}

	registry.Register(customQueueSerializer{})
	customPayload, err := registry.Encode("custom", envelope, CompressionNone)
	if err != nil {
		t.Fatalf("Encode(custom) error = %v", err)
	}
	var customDecoded Envelope
	if err := registry.Decode(customPayload, &customDecoded); err != nil {
		t.Fatalf("Decode(custom) error = %v", err)
	}
	if customDecoded.Name != envelope.Name {
		t.Fatalf("custom decoded = %#v", customDecoded)
	}

	if _, err := registry.Encode("json", envelope, CompressionZstd); !errors.Is(err, ErrUnsupportedCompression) {
		t.Fatalf("Encode(zstd) error = %v, want ErrUnsupportedCompression", err)
	}
}

func TestSignatureOptionsCountdownETAExpiresPriorityHeadersAndClone(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	signature := NewSignature("blog.publish", "go").
		WithKwarg("force", true).
		WithHeader("trace", "abc").
		WithQueue("emails").
		WithPriority(8).
		WithCountdown(5*time.Minute, now).
		WithExpires(expires)

	if signature.Options.ETA == nil || !signature.Options.ETA.Equal(now.Add(5*time.Minute)) || signature.Options.Expires == nil || !signature.Options.Expires.Equal(expires) {
		t.Fatalf("signature time options = %#v", signature.Options)
	}
	if signature.Options.Queue != "emails" || signature.Options.Priority != 8 || signature.Headers["trace"] != "abc" || signature.Kwargs["force"] != true {
		t.Fatalf("signature options = %#v", signature)
	}

	clone := signature.Clone()
	clone.Args[0] = "changed"
	clone.Kwargs["force"] = false
	clone.Headers["trace"] = "changed"
	if signature.Args[0] != "go" || signature.Kwargs["force"] != true || signature.Headers["trace"] != "abc" {
		t.Fatalf("Clone() leaked mutation: original=%#v clone=%#v", signature, clone)
	}

	eta := now.Add(10 * time.Minute)
	withETA := signature.WithETA(eta)
	if withETA.Options.ETA == nil || !withETA.Options.ETA.Equal(eta) {
		t.Fatalf("WithETA() = %#v", withETA.Options.ETA)
	}
	if reflect.DeepEqual(signature.Options.ETA, withETA.Options.ETA) {
		t.Fatalf("WithETA should be immutable")
	}
}

type customQueueSerializer struct{}

func (customQueueSerializer) Name() string        { return "custom" }
func (customQueueSerializer) ContentType() string { return "application/x-custom" }
func (customQueueSerializer) Trusted() bool       { return true }
func (customQueueSerializer) Marshal(value any) ([]byte, error) {
	return []byte(value.(Envelope).Name), nil
}
func (customQueueSerializer) Unmarshal(data []byte, value any) error {
	target := value.(*Envelope)
	target.Name = string(data)
	return nil
}
