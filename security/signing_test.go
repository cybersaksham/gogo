package security

import (
	"errors"
	"testing"
	"time"
)

func TestSignerValidTamperExpiryWrongSaltAndRotatedKeys(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	signer := NewSigner(SignerOptions{SecretKey: "current", Salt: "forms", Now: func() time.Time { return now }})
	token, err := signer.Sign("value")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	value, err := signer.Unsign(token, time.Minute)
	if err != nil || value != "value" {
		t.Fatalf("Unsign(valid) = %q, %v", value, err)
	}
	if _, err := signer.Unsign(token+"x", time.Minute); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("Unsign(tampered) error = %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := signer.Unsign(token, time.Minute); !errors.Is(err, ErrSignatureExpired) {
		t.Fatalf("Unsign(expired) error = %v", err)
	}
	now = now.Add(-2 * time.Minute)
	wrongSalt := NewSigner(SignerOptions{SecretKey: "current", Salt: "other", Now: func() time.Time { return now }})
	if _, err := wrongSalt.Unsign(token, time.Minute); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("Unsign(wrong salt) error = %v", err)
	}
	rotated := NewSigner(SignerOptions{SecretKey: "new", FallbackKeys: []string{"current"}, Salt: "forms", Now: func() time.Time { return now }})
	value, err = rotated.Unsign(token, time.Minute)
	if err != nil || value != "value" {
		t.Fatalf("Unsign(rotated) = %q, %v", value, err)
	}
}
