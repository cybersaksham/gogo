package fields

import "fmt"

// BinaryField stores byte slices.
type BinaryField struct {
	*BaseField
}

// NewBinaryField creates a binary field.
func NewBinaryField(options Options) *BinaryField {
	return &BinaryField{BaseField: NewBaseField("binary", options, map[string]string{"postgres": "bytea", "sqlite": "blob"})}
}

func (f *BinaryField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	if _, ok := value.([]byte); !ok {
		return fmt.Errorf("%w: %s must be bytes", ErrValidation, f.Name())
	}
	return nil
}

func (f *BinaryField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return append([]byte(nil), value.([]byte)...), nil
}

func (f *BinaryField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *BinaryField) Clone() Field {
	return &BinaryField{BaseField: f.BaseField.Clone().(*BaseField)}
}
