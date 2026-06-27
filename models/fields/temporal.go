package fields

import (
	"fmt"
	"time"
)

// TemporalConfig configures temporal field behavior.
type TemporalConfig struct {
	AutoNow    bool
	AutoNowAdd bool
	Location   *time.Location
	Now        func() time.Time
}

// DateField stores date-only values.
type DateField struct {
	*BaseField
	config TemporalConfig
}

// NewDateField creates a date field.
func NewDateField(options Options, configs ...TemporalConfig) *DateField {
	return &DateField{BaseField: NewBaseField("date", options, map[string]string{"postgres": "date", "sqlite": "date"}), config: temporalConfig(configs)}
}

func (f *DateField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	date, ok := value.(time.Time)
	if !ok || date.Hour() != 0 || date.Minute() != 0 || date.Second() != 0 || date.Nanosecond() != 0 {
		return fmt.Errorf("%w: %s must be date-only", ErrValidation, f.Name())
	}
	return nil
}

func (f *DateField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	date := value.(time.Time).In(f.location())
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, f.location()), nil
}

func (f *DateField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *DateField) AutoValue(isCreate bool) (time.Time, bool) {
	value, ok := autoTemporalValue(f.config, isCreate)
	if !ok {
		return time.Time{}, false
	}
	value = value.In(f.location())
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, f.location()), true
}

func (f *DateField) Clone() Field {
	return &DateField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}

func (f *DateField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

func (f *DateField) location() *time.Location {
	return temporalLocation(f.config)
}

// DateTimeField stores timestamp values.
type DateTimeField struct {
	*BaseField
	config TemporalConfig
}

// NewDateTimeField creates a datetime field.
func NewDateTimeField(options Options, configs ...TemporalConfig) *DateTimeField {
	return &DateTimeField{BaseField: NewBaseField("datetime", options, map[string]string{"postgres": "timestamptz", "sqlite": "datetime"}), config: temporalConfig(configs)}
}

func (f *DateTimeField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	if _, ok := value.(time.Time); !ok {
		return fmt.Errorf("%w: %s must be datetime", ErrValidation, f.Name())
	}
	return nil
}

func (f *DateTimeField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return value.(time.Time).In(temporalLocation(f.config)), nil
}

func (f *DateTimeField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *DateTimeField) AutoValue(isCreate bool) (time.Time, bool) {
	value, ok := autoTemporalValue(f.config, isCreate)
	if !ok {
		return time.Time{}, false
	}
	return value.In(temporalLocation(f.config)), true
}

func (f *DateTimeField) Clone() Field {
	return &DateTimeField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}

func (f *DateTimeField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

// TimeField stores time-only values.
type TimeField struct {
	*BaseField
	config TemporalConfig
}

// NewTimeField creates a time field.
func NewTimeField(options Options, configs ...TemporalConfig) *TimeField {
	return &TimeField{BaseField: NewBaseField("time", options, map[string]string{"postgres": "time", "sqlite": "time"}), config: temporalConfig(configs)}
}

func (f *TimeField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	timeValue, ok := value.(time.Time)
	if !ok || timeValue.Year() != 0 || timeValue.Month() != time.January || timeValue.Day() != 1 {
		return fmt.Errorf("%w: %s must be time-only", ErrValidation, f.Name())
	}
	return nil
}

func (f *TimeField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return value.(time.Time), nil
}

func (f *TimeField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *TimeField) AutoValue(isCreate bool) (time.Time, bool) {
	value, ok := autoTemporalValue(f.config, isCreate)
	if !ok {
		return time.Time{}, false
	}
	value = value.In(temporalLocation(f.config))
	return time.Date(0, 1, 1, value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), temporalLocation(f.config)), true
}

func (f *TimeField) Clone() Field {
	return &TimeField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}

func (f *TimeField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

// DurationField stores duration values.
type DurationField struct {
	*BaseField
}

// NewDurationField creates a duration field.
func NewDurationField(options Options) *DurationField {
	return &DurationField{BaseField: NewBaseField("duration", options, map[string]string{"postgres": "bigint", "sqlite": "integer"})}
}

func (f *DurationField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	if _, ok := value.(time.Duration); !ok {
		return fmt.Errorf("%w: %s must be duration", ErrValidation, f.Name())
	}
	return nil
}

func (f *DurationField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return int64(value.(time.Duration).Nanoseconds()), nil
}

func (f *DurationField) FromDB(value any) (any, error) {
	switch typed := value.(type) {
	case time.Duration:
		return typed, nil
	case int64:
		return time.Duration(typed), nil
	default:
		return nil, fmt.Errorf("%w: %s must be duration", ErrValidation, f.Name())
	}
}

func (f *DurationField) Clone() Field {
	return &DurationField{BaseField: f.BaseField.Clone().(*BaseField)}
}

func temporalConfig(configs []TemporalConfig) TemporalConfig {
	if len(configs) == 0 {
		return TemporalConfig{}
	}
	return configs[0]
}

func temporalLocation(config TemporalConfig) *time.Location {
	if config.Location != nil {
		return config.Location
	}
	return time.UTC
}

func autoTemporalValue(config TemporalConfig, isCreate bool) (time.Time, bool) {
	if !config.AutoNow && !(config.AutoNowAdd && isCreate) {
		return time.Time{}, false
	}
	now := time.Now
	if config.Now != nil {
		now = config.Now
	}
	return now(), true
}
