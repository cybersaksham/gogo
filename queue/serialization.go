package queue

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrUnknownSerializer      = errors.New("unknown serializer")
	ErrUntrustedSerializer    = errors.New("untrusted serializer")
	ErrUnsupportedCompression = errors.New("unsupported compression")
)

type Compression string

const (
	CompressionNone Compression = "none"
	CompressionGzip Compression = "gzip"
	CompressionZstd Compression = "zstd"
)

// Payload is a serialized queue message body.
type Payload struct {
	Serializer  string
	ContentType string
	Compression Compression
	Body        []byte
}

// Serializer encodes and decodes queue payloads.
type Serializer interface {
	Name() string
	ContentType() string
	Trusted() bool
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}

type SerializationOptions struct {
	AllowUntrustedSerializers []string
}

// SerializationRegistry stores serializers and trust settings.
type SerializationRegistry struct {
	serializers      map[string]Serializer
	allowedUntrusted map[string]bool
}

func NewSerializationRegistry(options SerializationOptions) *SerializationRegistry {
	registry := &SerializationRegistry{
		serializers:      map[string]Serializer{},
		allowedUntrusted: map[string]bool{},
	}
	for _, name := range options.AllowUntrustedSerializers {
		registry.allowedUntrusted[strings.ToLower(name)] = true
	}
	registry.Register(jsonSerializer{})
	registry.Register(gobSerializer{})
	registry.Register(rawSerializer{})
	return registry
}

func (r *SerializationRegistry) Register(serializer Serializer) {
	if serializer == nil {
		return
	}
	r.serializers[strings.ToLower(serializer.Name())] = serializer
}

func (r *SerializationRegistry) Encode(serializerName string, value any, compression Compression) (Payload, error) {
	serializer, err := r.serializer(serializerName)
	if err != nil {
		return Payload{}, err
	}
	if err := r.validateTrust(serializer); err != nil {
		return Payload{}, err
	}
	body, err := serializer.Marshal(value)
	if err != nil {
		return Payload{}, err
	}
	body, err = compressPayload(body, compression)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Serializer: serializer.Name(), ContentType: serializer.ContentType(), Compression: normalizedCompression(compression), Body: body}, nil
}

func (r *SerializationRegistry) Decode(payload Payload, value any) error {
	serializer, err := r.serializer(payload.Serializer)
	if err != nil {
		return err
	}
	if err := r.validateTrust(serializer); err != nil {
		return err
	}
	body, err := decompressPayload(payload.Body, payload.Compression)
	if err != nil {
		return err
	}
	return serializer.Unmarshal(body, value)
}

func (r *SerializationRegistry) serializer(name string) (Serializer, error) {
	serializer, ok := r.serializers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownSerializer, name)
	}
	return serializer, nil
}

func (r *SerializationRegistry) validateTrust(serializer Serializer) error {
	if serializer.Trusted() || r.allowedUntrusted[strings.ToLower(serializer.Name())] {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUntrustedSerializer, serializer.Name())
}

func compressPayload(body []byte, compression Compression) ([]byte, error) {
	switch normalizedCompression(compression) {
	case CompressionNone:
		return body, nil
	case CompressionGzip:
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		if _, err := writer.Write(body); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		return buffer.Bytes(), nil
	case CompressionZstd:
		return nil, fmt.Errorf("%w: zstd", ErrUnsupportedCompression)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCompression, compression)
	}
}

func decompressPayload(body []byte, compression Compression) ([]byte, error) {
	switch normalizedCompression(compression) {
	case CompressionNone:
		return body, nil
	case CompressionGzip:
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	case CompressionZstd:
		return nil, fmt.Errorf("%w: zstd", ErrUnsupportedCompression)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCompression, compression)
	}
}

func normalizedCompression(compression Compression) Compression {
	if compression == "" {
		return CompressionNone
	}
	return compression
}

type jsonSerializer struct{}

func (jsonSerializer) Name() string        { return "json" }
func (jsonSerializer) ContentType() string { return "application/json" }
func (jsonSerializer) Trusted() bool       { return true }
func (jsonSerializer) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}
func (jsonSerializer) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

type gobSerializer struct{}

func (gobSerializer) Name() string        { return "gob" }
func (gobSerializer) ContentType() string { return "application/gob" }
func (gobSerializer) Trusted() bool       { return false }
func (gobSerializer) Marshal(value any) ([]byte, error) {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
func (gobSerializer) Unmarshal(data []byte, value any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(value)
}

type rawSerializer struct{}

func (rawSerializer) Name() string        { return "raw" }
func (rawSerializer) ContentType() string { return "application/octet-stream" }
func (rawSerializer) Trusted() bool       { return true }
func (rawSerializer) Marshal(value any) ([]byte, error) {
	switch typed := value.(type) {
	case []byte:
		return append([]byte(nil), typed...), nil
	case string:
		return []byte(typed), nil
	default:
		return nil, fmt.Errorf("%w: raw serializer expects bytes", ErrInvalidTask)
	}
}
func (rawSerializer) Unmarshal(data []byte, value any) error {
	switch target := value.(type) {
	case *[]byte:
		*target = append([]byte(nil), data...)
		return nil
	case *string:
		*target = string(data)
		return nil
	default:
		return fmt.Errorf("%w: raw serializer target must be bytes or string", ErrInvalidTask)
	}
}

func init() {
	gob.Register(Envelope{})
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register(map[string]string{})
}
