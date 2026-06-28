package queue

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMessageSigningKey       = errors.New("message signing key error")
	ErrInvalidMessageSignature = errors.New("invalid message signature")
	ErrMessageExpired          = errors.New("message timestamp outside replay window")
	ErrRejectedContentType     = errors.New("rejected content type")
	ErrInvalidBrokerTLS        = errors.New("invalid broker tls configuration")
)

const (
	SignatureHeader = "x-gogo-signature"
	TimestampHeader = "x-gogo-timestamp"
	KeyIDHeader     = "x-gogo-key-id"
	RedactedValue   = "[REDACTED]"
)

type MessageSignerOptions struct {
	PrimaryKeyID string
	Keys         map[string][]byte
	Now          func() time.Time
	ReplayWindow time.Duration
}

type MessageSigner struct {
	primaryKeyID string
	keys         map[string][]byte
	now          func() time.Time
	replayWindow time.Duration
}

func NewMessageSigner(options MessageSignerOptions) *MessageSigner {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	keys := make(map[string][]byte, len(options.Keys))
	for id, key := range options.Keys {
		keys[id] = append([]byte(nil), key...)
	}
	if options.PrimaryKeyID == "" {
		for id := range keys {
			options.PrimaryKeyID = id
			break
		}
	}
	if options.ReplayWindow == 0 {
		options.ReplayWindow = 5 * time.Minute
	}
	return &MessageSigner{primaryKeyID: options.PrimaryKeyID, keys: keys, now: now, replayWindow: options.ReplayWindow}
}

func (s *MessageSigner) Sign(envelope Envelope) (map[string]string, error) {
	key, ok := s.keys[s.primaryKeyID]
	if !ok || len(key) == 0 {
		return nil, fmt.Errorf("%w: primary key missing", ErrMessageSigningKey)
	}
	timestamp := s.now().UTC().Unix()
	signature, err := s.signature(envelope, timestamp, key)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		KeyIDHeader:     s.primaryKeyID,
		TimestampHeader: strconv.FormatInt(timestamp, 10),
		SignatureHeader: signature,
	}, nil
}

func (s *MessageSigner) Verify(envelope Envelope, headers map[string]string) error {
	keyID := headers[KeyIDHeader]
	key, ok := s.keys[keyID]
	if !ok || len(key) == 0 {
		return fmt.Errorf("%w: key %q", ErrMessageSigningKey, keyID)
	}
	timestamp, err := strconv.ParseInt(headers[TimestampHeader], 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid timestamp", ErrInvalidMessageSignature)
	}
	if s.replayWindow > 0 {
		signedAt := time.Unix(timestamp, 0).UTC()
		now := s.now().UTC()
		if now.Sub(signedAt) > s.replayWindow || signedAt.Sub(now) > s.replayWindow {
			return ErrMessageExpired
		}
	}
	expected, err := s.signature(envelope, timestamp, key)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(expected), []byte(headers[SignatureHeader])) {
		return ErrInvalidMessageSignature
	}
	return nil
}

func (s *MessageSigner) signature(envelope Envelope, timestamp int64, key []byte) (string, error) {
	body, err := canonicalEnvelope(envelope)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	return base64.RawStdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func canonicalEnvelope(envelope Envelope) ([]byte, error) {
	envelope.Headers = withoutSecurityHeaders(envelope.Headers)
	return json.Marshal(envelope)
}

func withoutSecurityHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	filtered := make(map[string]string, len(headers))
	for key, value := range headers {
		lower := strings.ToLower(key)
		if lower == SignatureHeader || lower == TimestampHeader || lower == KeyIDHeader {
			continue
		}
		filtered[key] = value
	}
	return filtered
}

type ContentTypeAllowlist struct {
	accepted map[string]struct{}
}

func NewContentTypeAllowlist(contentTypes ...string) ContentTypeAllowlist {
	accepted := make(map[string]struct{}, len(contentTypes))
	for _, contentType := range contentTypes {
		accepted[strings.ToLower(strings.TrimSpace(contentType))] = struct{}{}
	}
	return ContentTypeAllowlist{accepted: accepted}
}

func (a ContentTypeAllowlist) Validate(contentType string) error {
	if len(a.accepted) == 0 {
		return nil
	}
	if _, ok := a.accepted[strings.ToLower(strings.TrimSpace(contentType))]; !ok {
		return fmt.Errorf("%w: %s", ErrRejectedContentType, contentType)
	}
	return nil
}

type BrokerTLSConfig struct {
	URL                string
	TLSEnabled         bool
	ServerName         string
	InsecureSkipVerify bool
}

func ValidateBrokerTLS(config BrokerTLSConfig) error {
	if config.InsecureSkipVerify {
		return fmt.Errorf("%w: insecure skip verify is not allowed", ErrInvalidBrokerTLS)
	}
	parsed, err := url.Parse(config.URL)
	if err != nil && config.URL != "" {
		return fmt.Errorf("%w: %v", ErrInvalidBrokerTLS, err)
	}
	secureScheme := parsed.Scheme == "amqps" || parsed.Scheme == "rediss" || parsed.Scheme == "https"
	if secureScheme && !config.TLSEnabled {
		return fmt.Errorf("%w: secure broker URL requires TLS", ErrInvalidBrokerTLS)
	}
	if config.TLSEnabled && parsed.Host != "" && config.ServerName == "" {
		return fmt.Errorf("%w: server name is required", ErrInvalidBrokerTLS)
	}
	return nil
}

type SensitiveValue struct {
	Value any
}

func Sensitive(value any) SensitiveValue {
	return SensitiveValue{Value: value}
}

type RedactorOptions struct {
	SensitiveKeys []string
}

type Redactor struct {
	keys map[string]struct{}
}

func NewRedactor(options RedactorOptions) Redactor {
	keys := map[string]struct{}{}
	defaults := []string{"password", "passwd", "secret", "token", "access_token", "refresh_token", "api_key", "authorization", "cookie"}
	for _, key := range defaults {
		keys[key] = struct{}{}
	}
	for _, key := range options.SensitiveKeys {
		keys[strings.ToLower(key)] = struct{}{}
	}
	return Redactor{keys: keys}
}

func (r Redactor) RedactEnvelope(envelope Envelope) Envelope {
	envelope.Args = r.redactSlice(envelope.Args)
	envelope.Kwargs = r.redactMap(envelope.Kwargs)
	return envelope
}

func (r Redactor) RedactEvent(event Event) Event {
	event.Fields = r.redactMap(event.Fields)
	return event
}

func (r Redactor) RedactResult(result Result) Result {
	result.Result = r.redactValue("", result.Result)
	result.Error = redactString(result.Error)
	result.Traceback = redactString(result.Traceback)
	return result
}

func (r Redactor) redactSlice(values []any) []any {
	if values == nil {
		return nil
	}
	redacted := make([]any, len(values))
	for i, value := range values {
		redacted[i] = r.redactValue("", value)
	}
	return redacted
}

func (r Redactor) redactMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	redacted := make(map[string]any, len(values))
	for key, value := range values {
		redacted[key] = r.redactValue(key, value)
	}
	return redacted
}

func (r Redactor) redactValue(key string, value any) any {
	if _, ok := value.(SensitiveValue); ok {
		return RedactedValue
	}
	if r.isSensitiveKey(key) {
		return RedactedValue
	}
	switch typed := value.(type) {
	case map[string]any:
		return r.redactMap(typed)
	case []any:
		return r.redactSlice(typed)
	default:
		return value
	}
}

func (r Redactor) isSensitiveKey(key string) bool {
	if key == "" {
		return false
	}
	_, ok := r.keys[strings.ToLower(key)]
	return ok
}

func redactString(value string) string {
	if value == "" {
		return ""
	}
	return value
}
