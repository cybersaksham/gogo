package fields

import (
	"encoding/json"
	"fmt"
)

// JSONField stores JSON values.
type JSONField struct {
	*BaseField
}

// NewJSONField creates a JSON field.
func NewJSONField(options Options) *JSONField {
	return &JSONField{BaseField: NewBaseField("json", options, map[string]string{"postgres": "jsonb", "sqlite": "json"})}
}

func (f *JSONField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	if _, err := json.Marshal(value); err != nil {
		return fmt.Errorf("%w: %s must be JSON marshalable", ErrValidation, f.Name())
	}
	return nil
}

func (f *JSONField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func (f *JSONField) FromDB(value any) (any, error) {
	var raw json.RawMessage
	switch typed := value.(type) {
	case []byte:
		raw = append(json.RawMessage(nil), typed...)
	case string:
		raw = append(json.RawMessage(nil), typed...)
	case json.RawMessage:
		raw = append(json.RawMessage(nil), typed...)
	default:
		return nil, fmt.Errorf("%w: %s must scan from JSON bytes", ErrValidation, f.Name())
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("%w: %s scanned invalid JSON", ErrValidation, f.Name())
	}
	return raw, nil
}

func (f *JSONField) Clone() Field {
	return &JSONField{BaseField: f.BaseField.Clone().(*BaseField)}
}
