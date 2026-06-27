package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

// PasswordResetTokenSigner creates expiring, password-bound reset tokens.
type PasswordResetTokenSigner struct {
	Secret string
	MaxAge time.Duration
	Now    func() time.Time
}

type passwordResetPayload struct {
	UserID    int64  `json:"uid"`
	Timestamp int64  `json:"ts"`
	Hash      string `json:"hash"`
}

// MakeToken signs a password reset token for a user.
func (s PasswordResetTokenSigner) MakeToken(user User) (string, error) {
	now := s.now()
	payload := passwordResetPayload{
		UserID:    user.ID,
		Timestamp: now.Unix(),
		Hash:      s.userHash(user),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	value := base64.RawURLEncoding.EncodeToString(body)
	return value + "." + s.signature(value), nil
}

// CheckToken verifies token signature, expiry, user id, and password hash binding.
func (s PasswordResetTokenSigner) CheckToken(user User, token string) (bool, error) {
	value, sig, ok := strings.Cut(token, ".")
	if !ok || value == "" || sig == "" {
		return false, ErrInvalidPasswordResetToken
	}
	if subtle.ConstantTimeCompare([]byte(sig), []byte(s.signature(value))) != 1 {
		return false, ErrInvalidPasswordResetToken
	}
	body, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return false, ErrInvalidPasswordResetToken
	}
	var payload passwordResetPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, ErrInvalidPasswordResetToken
	}
	if payload.UserID != user.ID || payload.Hash != s.userHash(user) {
		return false, ErrInvalidPasswordResetToken
	}
	maxAge := s.MaxAge
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}
	issued := time.Unix(payload.Timestamp, 0).UTC()
	if s.now().Before(issued) || s.now().Sub(issued) > maxAge {
		return false, ErrInvalidPasswordResetToken
	}
	return true, nil
}

func (s PasswordResetTokenSigner) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s PasswordResetTokenSigner) signature(value string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s PasswordResetTokenSigner) userHash(user User) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte(strconv.FormatInt(user.ID, 10)))
	mac.Write([]byte("|"))
	mac.Write([]byte(user.Password))
	mac.Write([]byte("|"))
	mac.Write([]byte(strconv.FormatBool(user.IsActive)))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
