package fields

import (
	"fmt"
	"reflect"
)

// Validator validates one field value.
type Validator func(any) error

// Choice describes one selectable field value.
type Choice struct {
	Value any
	Label string
	Group string
}

// Options contains common Django-style field options.
type Options struct {
	Name           string
	Column         string
	PrimaryKey     bool
	Unique         bool
	DBIndex        bool
	DBDefault      any
	DBCollation    string
	Null           bool
	Blank          bool
	Default        any
	Choices        []Choice
	Editable       *bool
	HelpText       string
	VerboseName    string
	ErrorMessages  map[string]string
	Validators     []Validator
	UniqueForDate  string
	UniqueForMonth string
	UniqueForYear  string
	Serialize      *bool
	DBComment      string
	DBTablespace   string
}

// Field is the common model field contract.
type Field interface {
	Name() string
	ColumnName() string
	Options() Options
	Kind() string
	ColumnType(dialect string) string
	Validate(value any) error
	ToDB(value any) (any, error)
	FromDB(value any) (any, error)
	Clone() Field
}

// BaseField implements common field behavior.
type BaseField struct {
	kind        string
	options     Options
	columnTypes map[string]string
}

// NewBaseField creates a field with common behavior.
func NewBaseField(kind string, options Options, columnTypes map[string]string) *BaseField {
	return &BaseField{
		kind:        kind,
		options:     options.clone(),
		columnTypes: cloneStringMap(columnTypes),
	}
}

// Name returns the field name.
func (f *BaseField) Name() string {
	return f.options.Name
}

// ColumnName returns the database column name.
func (f *BaseField) ColumnName() string {
	if f.options.Column != "" {
		return f.options.Column
	}
	return f.options.Name
}

// Options returns copied options.
func (f *BaseField) Options() Options {
	return f.options.clone()
}

// IsEditable returns whether admin/forms may edit this field.
func (f *BaseField) IsEditable() bool {
	if f.options.Editable == nil {
		return true
	}
	return *f.options.Editable
}

// IsSerializable returns whether serializers should include this field.
func (f *BaseField) IsSerializable() bool {
	if f.options.Serialize == nil {
		return true
	}
	return *f.options.Serialize
}

// Kind returns the field kind.
func (f *BaseField) Kind() string {
	return f.kind
}

// ColumnType returns a database column type for a dialect.
func (f *BaseField) ColumnType(dialect string) string {
	if f.columnTypes == nil {
		return f.kind
	}
	if columnType, ok := f.columnTypes[dialect]; ok {
		return columnType
	}
	if columnType, ok := f.columnTypes["default"]; ok {
		return columnType
	}
	return f.kind
}

// Validate validates common null, blank, choice, and custom validator behavior.
func (f *BaseField) Validate(value any) error {
	if value == nil {
		if f.options.Null {
			return nil
		}
		return fmt.Errorf("%w: %s cannot be null", ErrValidation, f.Name())
	}
	if value == "" {
		if f.options.Blank {
			return nil
		}
		return fmt.Errorf("%w: %s cannot be blank", ErrValidation, f.Name())
	}
	if len(f.options.Choices) > 0 && !f.validChoice(value) {
		return fmt.Errorf("%w: %s has invalid choice", ErrValidation, f.Name())
	}
	for _, validator := range f.options.Validators {
		if err := validator(value); err != nil {
			return fmt.Errorf("%w: %v", ErrValidation, err)
		}
	}
	return nil
}

// ToDB converts a value for database storage.
func (f *BaseField) ToDB(value any) (any, error) {
	return value, nil
}

// FromDB converts a database value for model use.
func (f *BaseField) FromDB(value any) (any, error) {
	return value, nil
}

// Clone returns an independent field copy.
func (f *BaseField) Clone() Field {
	return &BaseField{
		kind:        f.kind,
		options:     f.options.clone(),
		columnTypes: cloneStringMap(f.columnTypes),
	}
}

func (f *BaseField) validChoice(value any) bool {
	for _, choice := range f.options.Choices {
		if reflect.DeepEqual(choice.Value, value) {
			return true
		}
	}
	return false
}

func (o Options) clone() Options {
	copied := o
	copied.Choices = append([]Choice(nil), o.Choices...)
	copied.ErrorMessages = cloneStringMap(o.ErrorMessages)
	copied.Validators = append([]Validator(nil), o.Validators...)
	if o.Editable != nil {
		value := *o.Editable
		copied.Editable = &value
	}
	if o.Serialize != nil {
		value := *o.Serialize
		copied.Serialize = &value
	}
	return copied
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
