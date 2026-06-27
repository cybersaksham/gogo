package fields

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// PostgresField marks fields that require PostgreSQL capabilities.
type PostgresField interface {
	Field
	RequireDialect(string) error
}

// ArrayField stores PostgreSQL arrays.
type ArrayField struct {
	*BaseField
	element Field
}

// NewArrayField creates a PostgreSQL array field.
func NewArrayField(options Options, element Field) *ArrayField {
	return &ArrayField{BaseField: NewBaseField("array", options, nil), element: element}
}

func (f *ArrayField) ColumnType(dialect string) string {
	if dialect == "postgres" || dialect == "postgresql" {
		return f.element.ColumnType("postgres") + "[]"
	}
	return "array"
}

func (f *ArrayField) RequireDialect(dialect string) error {
	return requirePostgresDialect(dialect, f.Name())
}

func (f *ArrayField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	kind := reflect.TypeOf(value).Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return fmt.Errorf("%w: %s must be array", ErrValidation, f.Name())
	}
	return nil
}

func (f *ArrayField) Clone() Field {
	return &ArrayField{BaseField: f.BaseField.Clone().(*BaseField), element: f.element.Clone()}
}

// HStoreField stores PostgreSQL hstore-compatible string maps.
type HStoreField struct {
	*BaseField
}

// NewHStoreField creates an hstore field.
func NewHStoreField(options Options) *HStoreField {
	return &HStoreField{BaseField: NewBaseField("hstore", options, map[string]string{"postgres": "hstore"})}
}

func (f *HStoreField) RequireDialect(dialect string) error {
	return requirePostgresDialect(dialect, f.Name())
}

func (f *HStoreField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	if _, ok := value.(map[string]string); !ok {
		return fmt.Errorf("%w: %s must be map[string]string", ErrValidation, f.Name())
	}
	return nil
}

func (f *HStoreField) Clone() Field {
	return &HStoreField{BaseField: f.BaseField.Clone().(*BaseField)}
}

// Range stores a PostgreSQL range value.
type Range struct {
	Lower  any
	Upper  any
	Bounds string
}

// RangeField stores PostgreSQL range metadata.
type RangeField struct {
	*BaseField
	columnType string
	validate   func(Range) error
}

// NewIntegerRangeField creates an int4range field.
func NewIntegerRangeField(options Options) *RangeField {
	return newRangeField("integer_range", options, "int4range", validateIntRange)
}

// NewBigIntegerRangeField creates an int8range field.
func NewBigIntegerRangeField(options Options) *RangeField {
	return newRangeField("big_integer_range", options, "int8range", validateIntRange)
}

// NewDecimalRangeField creates a numrange field.
func NewDecimalRangeField(options Options) *RangeField {
	return newRangeField("decimal_range", options, "numrange", validateDecimalRange)
}

// NewDateRangeField creates a daterange field.
func NewDateRangeField(options Options) *RangeField {
	return newRangeField("date_range", options, "daterange", validateTimeRange)
}

// NewDateTimeRangeField creates a tstzrange field.
func NewDateTimeRangeField(options Options) *RangeField {
	return newRangeField("datetime_range", options, "tstzrange", validateTimeRange)
}

func newRangeField(kind string, options Options, columnType string, validator func(Range) error) *RangeField {
	return &RangeField{BaseField: NewBaseField(kind, options, map[string]string{"postgres": columnType}), columnType: columnType, validate: validator}
}

func (f *RangeField) RequireDialect(dialect string) error {
	return requirePostgresDialect(dialect, f.Name())
}

func (f *RangeField) ColumnType(dialect string) string {
	if dialect == "postgres" || dialect == "postgresql" {
		return f.columnType
	}
	return f.Kind()
}

func (f *RangeField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	rangeValue, ok := value.(Range)
	if !ok {
		return fmt.Errorf("%w: %s must be range", ErrValidation, f.Name())
	}
	if rangeValue.Bounds == "" {
		rangeValue.Bounds = "[)"
	}
	if err := f.validate(rangeValue); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (f *RangeField) Clone() Field {
	return &RangeField{BaseField: f.BaseField.Clone().(*BaseField), columnType: f.columnType, validate: f.validate}
}

func requirePostgresDialect(dialect, name string) error {
	if dialect == "postgres" || dialect == "postgresql" {
		return nil
	}
	return fmt.Errorf("%w: %s requires PostgreSQL", ErrInvalidField, name)
}

func validateIntRange(value Range) error {
	lower, err := toInt64(value.Lower)
	if err != nil {
		return err
	}
	upper, err := toInt64(value.Upper)
	if err != nil {
		return err
	}
	if lower > upper {
		return fmt.Errorf("lower bound is greater than upper bound")
	}
	return nil
}

func validateDecimalRange(value Range) error {
	lower, err := strconv.ParseFloat(fmt.Sprint(value.Lower), 64)
	if err != nil {
		return err
	}
	upper, err := strconv.ParseFloat(fmt.Sprint(value.Upper), 64)
	if err != nil {
		return err
	}
	if lower > upper {
		return fmt.Errorf("lower bound is greater than upper bound")
	}
	return nil
}

func validateTimeRange(value Range) error {
	lower, ok := value.Lower.(time.Time)
	if !ok {
		return fmt.Errorf("lower bound must be time")
	}
	upper, ok := value.Upper.(time.Time)
	if !ok {
		return fmt.Errorf("upper bound must be time")
	}
	if lower.After(upper) {
		return fmt.Errorf("lower bound is greater than upper bound")
	}
	return nil
}
