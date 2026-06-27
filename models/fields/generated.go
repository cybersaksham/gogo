package fields

import "fmt"

// GeneratedField describes a database-generated expression field.
type GeneratedField struct {
	*BaseField
	expression string
}

// NewGeneratedField creates a generated field.
func NewGeneratedField(options Options, expression string) *GeneratedField {
	editable := false
	options.Editable = &editable
	return &GeneratedField{BaseField: NewBaseField("generated", options, map[string]string{"postgres": "generated", "sqlite": "generated"}), expression: expression}
}

// Expression returns the generation expression.
func (f *GeneratedField) Expression() string {
	return f.expression
}

func (f *GeneratedField) ToDB(any) (any, error) {
	return nil, fmt.Errorf("%w: generated field %s cannot be manually saved", ErrInvalidField, f.Name())
}

func (f *GeneratedField) Clone() Field {
	return &GeneratedField{BaseField: f.BaseField.Clone().(*BaseField), expression: f.expression}
}
