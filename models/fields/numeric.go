package fields

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type integerField struct {
	*BaseField
	min int64
	max int64
}

// NewAutoField creates a 32-bit auto-incrementing primary key field.
func NewAutoField(options Options) Field {
	options.PrimaryKey = true
	return newIntegerField("auto", options, 0, math.MaxInt32, map[string]string{"postgres": "serial", "sqlite": "integer"})
}

// NewBigAutoField creates a 64-bit auto-incrementing primary key field.
func NewBigAutoField(options Options) Field {
	options.PrimaryKey = true
	return newIntegerField("big_auto", options, 0, math.MaxInt64, map[string]string{"postgres": "bigserial", "sqlite": "integer"})
}

// NewSmallAutoField creates a 16-bit auto-incrementing primary key field.
func NewSmallAutoField(options Options) Field {
	options.PrimaryKey = true
	return newIntegerField("small_auto", options, 0, math.MaxInt16, map[string]string{"postgres": "smallserial", "sqlite": "integer"})
}

// NewIntegerField creates a signed 32-bit integer field.
func NewIntegerField(options Options) Field {
	return newIntegerField("integer", options, math.MinInt32, math.MaxInt32, map[string]string{"postgres": "integer", "sqlite": "integer"})
}

// NewBigIntegerField creates a signed 64-bit integer field.
func NewBigIntegerField(options Options) Field {
	return newIntegerField("big_integer", options, math.MinInt64, math.MaxInt64, map[string]string{"postgres": "bigint", "sqlite": "integer"})
}

// NewSmallIntegerField creates a signed 16-bit integer field.
func NewSmallIntegerField(options Options) Field {
	return newIntegerField("small_integer", options, math.MinInt16, math.MaxInt16, map[string]string{"postgres": "smallint", "sqlite": "integer"})
}

// NewPositiveIntegerField creates an unsigned 32-bit integer field.
func NewPositiveIntegerField(options Options) Field {
	return newIntegerField("positive_integer", options, 0, math.MaxInt32, map[string]string{"postgres": "integer", "sqlite": "integer"})
}

// NewPositiveBigIntegerField creates an unsigned 64-bit integer field.
func NewPositiveBigIntegerField(options Options) Field {
	return newIntegerField("positive_big_integer", options, 0, math.MaxInt64, map[string]string{"postgres": "bigint", "sqlite": "integer"})
}

// NewPositiveSmallIntegerField creates an unsigned 16-bit integer field.
func NewPositiveSmallIntegerField(options Options) Field {
	return newIntegerField("positive_small_integer", options, 0, math.MaxInt16, map[string]string{"postgres": "smallint", "sqlite": "integer"})
}

func newIntegerField(kind string, options Options, min, max int64, columnTypes map[string]string) Field {
	return &integerField{BaseField: NewBaseField(kind, options, columnTypes), min: min, max: max}
}

func (f *integerField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	integer, err := toInt64(value)
	if err != nil {
		return fmt.Errorf("%w: %s must be an integer", ErrValidation, f.Name())
	}
	if integer < f.min || integer > f.max {
		return fmt.Errorf("%w: %s out of range", ErrValidation, f.Name())
	}
	return nil
}

func (f *integerField) ToDB(value any) (any, error) {
	integer, err := toInt64(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return integer, nil
}

func (f *integerField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *integerField) Clone() Field {
	return &integerField{BaseField: f.BaseField.Clone().(*BaseField), min: f.min, max: f.max}
}

func (f *integerField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

// DecimalField stores fixed-precision decimal values.
type DecimalField struct {
	*BaseField
	maxDigits     int
	decimalPlaces int
}

// NewDecimalField creates a decimal field.
func NewDecimalField(options Options, maxDigits, decimalPlaces int) *DecimalField {
	return &DecimalField{
		BaseField:     NewBaseField("decimal", options, map[string]string{"postgres": fmt.Sprintf("numeric(%d,%d)", maxDigits, decimalPlaces), "sqlite": "text"}),
		maxDigits:     maxDigits,
		decimalPlaces: decimalPlaces,
	}
}

func (f *DecimalField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	digits, places, ok := decimalShape(fmt.Sprint(value))
	if !ok {
		return fmt.Errorf("%w: %s must be decimal", ErrValidation, f.Name())
	}
	if digits > f.maxDigits {
		return fmt.Errorf("%w: %s exceeds max digits", ErrValidation, f.Name())
	}
	if places > f.decimalPlaces {
		return fmt.Errorf("%w: %s exceeds decimal places", ErrValidation, f.Name())
	}
	return nil
}

func (f *DecimalField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return fmt.Sprint(value), nil
}

func (f *DecimalField) FromDB(value any) (any, error) {
	return fmt.Sprint(value), nil
}

func (f *DecimalField) Clone() Field {
	return &DecimalField{
		BaseField:     f.BaseField.Clone().(*BaseField),
		maxDigits:     f.maxDigits,
		decimalPlaces: f.decimalPlaces,
	}
}

func (f *DecimalField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

// FloatField stores floating point values.
type FloatField struct {
	*BaseField
}

// NewFloatField creates a floating point field.
func NewFloatField(options Options) *FloatField {
	return &FloatField{BaseField: NewBaseField("float", options, map[string]string{"postgres": "double precision", "sqlite": "real"})}
}

func (f *FloatField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null || value == "" && f.options.Blank {
		return err
	}
	if _, err := toFloat64(value); err != nil {
		return fmt.Errorf("%w: %s must be a float", ErrValidation, f.Name())
	}
	return nil
}

func (f *FloatField) ToDB(value any) (any, error) {
	float, err := toFloat64(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return float, nil
}

func (f *FloatField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *FloatField) Clone() Field {
	return &FloatField{BaseField: f.BaseField.Clone().(*BaseField)}
}

func toInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int8:
		return int64(typed), nil
	case int16:
		return int64(typed), nil
	case int32:
		return int64(typed), nil
	case int64:
		return typed, nil
	case uint:
		return uintToInt64(uint64(typed))
	case uint8:
		return int64(typed), nil
	case uint16:
		return int64(typed), nil
	case uint32:
		return int64(typed), nil
	case uint64:
		return uintToInt64(typed)
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported integer type %T", value)
	}
}

func uintToInt64(value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("integer overflows int64")
	}
	return int64(value), nil
}

func toFloat64(value any) (float64, error) {
	switch typed := value.(type) {
	case float32:
		return float64(typed), nil
	case float64:
		return typed, nil
	case int:
		return float64(typed), nil
	case int8:
		return float64(typed), nil
	case int16:
		return float64(typed), nil
	case int32:
		return float64(typed), nil
	case int64:
		return float64(typed), nil
	case uint:
		return float64(typed), nil
	case uint8:
		return float64(typed), nil
	case uint16:
		return float64(typed), nil
	case uint32:
		return float64(typed), nil
	case uint64:
		return float64(typed), nil
	case string:
		return strconv.ParseFloat(typed, 64)
	default:
		return 0, fmt.Errorf("unsupported float type %T", value)
	}
}

func decimalShape(value string) (int, int, bool) {
	value = strings.TrimPrefix(strings.TrimPrefix(value, "-"), "+")
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, 0, false
	}

	digits := 0
	for _, part := range parts {
		for _, char := range part {
			if char < '0' || char > '9' {
				return 0, 0, false
			}
			digits++
		}
	}
	places := 0
	if len(parts) == 2 {
		places = len(parts[1])
	}
	return digits, places, digits > 0
}
