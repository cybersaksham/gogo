package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrBadSignature     = errors.New("bad signature")
	ErrSignatureExpired = errors.New("signature expired")
	ErrMissingSecretKey = errors.New("missing secret key")
)

type SignerOptions struct {
	SecretKey    string
	FallbackKeys []string
	Salt         string
	Now          func() time.Time
}

type Signer struct {
	secretKey    string
	fallbackKeys []string
	salt         string
	now          func() time.Time
}

func NewSigner(options SignerOptions) *Signer {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &Signer{
		secretKey:    options.SecretKey,
		fallbackKeys: append([]string(nil), options.FallbackKeys...),
		salt:         options.Salt,
		now:          now,
	}
}

func (s *Signer) Sign(value string) (string, error) {
	if s.secretKey == "" {
		return "", ErrMissingSecretKey
	}
	timestamp := s.now().UTC().Unix()
	payload := base64.RawURLEncoding.EncodeToString([]byte(value))
	signature := s.signature(s.secretKey, payload, timestamp)
	return fmt.Sprintf("%s:%d:%s", payload, timestamp, signature), nil
}

func (s *Signer) Unsign(token string, maxAge time.Duration) (string, error) {
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return "", ErrBadSignature
	}
	payload := parts[0]
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", ErrBadSignature
	}
	if maxAge > 0 && s.now().UTC().Sub(time.Unix(timestamp, 0).UTC()) > maxAge {
		return "", ErrSignatureExpired
	}
	keys := append([]string{s.secretKey}, s.fallbackKeys...)
	for _, key := range keys {
		if key == "" {
			continue
		}
		expected := s.signature(key, payload, timestamp)
		if hmac.Equal([]byte(expected), []byte(parts[2])) {
			body, err := base64.RawURLEncoding.DecodeString(payload)
			if err != nil {
				return "", ErrBadSignature
			}
			return string(body), nil
		}
	}
	return "", ErrBadSignature
}

func (s *Signer) signature(secret string, payload string, timestamp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(s.salt))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(payload))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
