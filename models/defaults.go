package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// DefaultKind identifies how a database default should be rendered.
type DefaultKind string

const (
	DefaultNone       DefaultKind = ""
	DefaultLiteral    DefaultKind = "literal"
	DefaultExpression DefaultKind = "expression"
)

// DatabaseDefault stores a database-side column default.
//
// Literal defaults are safely quoted by schema renderers. Expression defaults
// are trusted SQL and should only be built from static application code.
type DatabaseDefault struct {
	Kind  DefaultKind `json:"kind"`
	Value any         `json:"value,omitempty"`
	SQL   string      `json:"sql,omitempty"`
}

// DefaultValue creates a literal database default.
func DefaultValue(value any) DatabaseDefault {
	return DatabaseDefault{Kind: DefaultLiteral, Value: value}
}

// DefaultSQL creates a trusted SQL expression database default.
func DefaultSQL(expression string) DatabaseDefault {
	return DatabaseDefault{Kind: DefaultExpression, SQL: strings.TrimSpace(expression)}
}

// NormalizeDatabaseDefault converts supported public default values into the
// canonical representation used by migrations.
func NormalizeDatabaseDefault(value any) (DatabaseDefault, error) {
	if value == nil {
		return DatabaseDefault{}, nil
	}
	switch typed := value.(type) {
	case DatabaseDefault:
		return normalizeDatabaseDefault(typed)
	case *DatabaseDefault:
		if typed == nil {
			return DatabaseDefault{}, nil
		}
		return normalizeDatabaseDefault(*typed)
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return DefaultValue(typed), nil
	default:
		return DatabaseDefault{}, fmt.Errorf("%w: unsupported database default %T", ErrInvalidMetadata, value)
	}
}

// UnmarshalJSON preserves compatibility with older manifests that stored
// db_default as a primitive JSON value.
func (d *DatabaseDefault) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) || len(trimmed) == 0 {
		*d = DatabaseDefault{}
		return nil
	}
	if len(trimmed) > 0 && trimmed[0] == '{' {
		var raw struct {
			Kind  DefaultKind     `json:"kind"`
			Value json.RawMessage `json:"value"`
			SQL   string          `json:"sql"`
		}
		if err := json.Unmarshal(trimmed, &raw); err != nil {
			return err
		}
		defaultValue := DatabaseDefault{Kind: raw.Kind, SQL: raw.SQL}
		if len(raw.Value) > 0 {
			value, err := decodeDefaultLiteral(raw.Value)
			if err != nil {
				return err
			}
			defaultValue.Value = value
		}
		normalized, err := normalizeDatabaseDefault(defaultValue)
		if err != nil {
			return err
		}
		*d = normalized
		return nil
	}
	value, err := decodeDefaultLiteral(trimmed)
	if err != nil {
		return err
	}
	*d = DefaultValue(value)
	return nil
}

func normalizeDatabaseDefault(defaultValue DatabaseDefault) (DatabaseDefault, error) {
	switch defaultValue.Kind {
	case DefaultNone:
		return DatabaseDefault{}, nil
	case DefaultLiteral:
		if _, err := NormalizeDatabaseDefault(defaultValue.Value); err != nil {
			return DatabaseDefault{}, err
		}
		return DatabaseDefault{Kind: DefaultLiteral, Value: normalizeDefaultNumber(defaultValue.Value)}, nil
	case DefaultExpression:
		expression := strings.TrimSpace(defaultValue.SQL)
		if expression == "" {
			return DatabaseDefault{}, fmt.Errorf("%w: database default expression cannot be empty", ErrInvalidMetadata)
		}
		return DatabaseDefault{Kind: DefaultExpression, SQL: expression}, nil
	default:
		return DatabaseDefault{}, fmt.Errorf("%w: unknown database default kind %q", ErrInvalidMetadata, defaultValue.Kind)
	}
}

func decodeDefaultLiteral(data []byte) (any, error) {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case nil, string, bool:
		return typed, nil
	case float64:
		if math.Trunc(typed) == typed {
			return int64(typed), nil
		}
		return typed, nil
	default:
		return nil, fmt.Errorf("%w: unsupported database default literal %T", ErrInvalidMetadata, value)
	}
}

func normalizeDefaultNumber(value any) any {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case uint:
		return uint64(typed)
	case uint8:
		return uint64(typed)
	case uint16:
		return uint64(typed)
	case uint32:
		return uint64(typed)
	case float32:
		return float64(typed)
	default:
		return value
	}
}
