package sessions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
)

const signedSeparator = "."

// NewSignedSessionKey creates an opaque signed server-side session key.
func NewSignedSessionKey(secret string) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	return signValue(secret, token), nil
}

// VerifySessionKey verifies an opaque signed server-side session key.
func VerifySessionKey(secret, key string) bool {
	_, ok := verifySignedValue(secret, key)
	return ok
}

func signValue(secret, value string) string {
	return value + signedSeparator + signature(secret, value)
}

func verifySignedValue(secret, signed string) (string, bool) {
	value, sig, ok := strings.Cut(signed, signedSeparator)
	if !ok || value == "" || sig == "" {
		return "", false
	}
	expected := signature(secret, value)
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) != 1 {
		return "", false
	}
	return value, true
}

func signature(secret, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
