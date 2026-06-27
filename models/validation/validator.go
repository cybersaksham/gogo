package validation

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	slugPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// Validator validates a value map and appends structured errors.
type Validator func(context.Context, map[string]any, *Errors)

// Validate runs validators and returns structured errors.
func Validate(ctx context.Context, values map[string]any, validators ...Validator) error {
	errs := &Errors{Fields: make(map[string][]Error)}
	for _, validator := range validators {
		if ctx.Err() != nil {
			errs.Add("__all__", "context", ctx.Err().Error())
			break
		}
		validator(ctx, values, errs)
	}
	if errs.empty() {
		return nil
	}
	return errs
}

// Required validates non-empty values.
func Required(field string) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		value := values[field]
		if value == nil || value == "" {
			errs.Add(field, "required", "This field is required.")
		}
	}
}

// Type validates the value has the same concrete type as example.
func Type(field string, example any) Validator {
	expected := reflect.TypeOf(example)
	return func(_ context.Context, values map[string]any, errs *Errors) {
		value := values[field]
		if value == nil || reflect.TypeOf(value) != expected {
			errs.Add(field, "type", fmt.Sprintf("Expected %s.", expected))
		}
	}
}

func MinValue(field string, min float64) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		value, ok := numeric(values[field])
		if !ok || value < min {
			errs.Add(field, "min_value", fmt.Sprintf("Ensure this value is at least %v.", min))
		}
	}
}

func MaxValue(field string, max float64) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		value, ok := numeric(values[field])
		if !ok || value > max {
			errs.Add(field, "max_value", fmt.Sprintf("Ensure this value is at most %v.", max))
		}
	}
}

func MinLength(field string, min int) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		length, ok := lengthOf(values[field])
		if !ok || length < min {
			errs.Add(field, "min_length", fmt.Sprintf("Ensure this value has at least %d characters.", min))
		}
	}
}

func MaxLength(field string, max int) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		length, ok := lengthOf(values[field])
		if !ok || length > max {
			errs.Add(field, "max_length", fmt.Sprintf("Ensure this value has at most %d characters.", max))
		}
	}
}

func Regex(field string, pattern *regexp.Regexp) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		if !pattern.MatchString(fmt.Sprint(values[field])) {
			errs.Add(field, "regex", "Enter a valid value.")
		}
	}
}

func Email(field string) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		value := fmt.Sprint(values[field])
		address, err := mail.ParseAddress(value)
		if err != nil || address.Address != value {
			errs.Add(field, "email", "Enter a valid email address.")
		}
	}
}

func URL(field string) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		parsed, err := url.Parse(fmt.Sprint(values[field]))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			errs.Add(field, "url", "Enter a valid URL.")
		}
	}
}

func Slug(field string) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		if !slugPattern.MatchString(fmt.Sprint(values[field])) {
			errs.Add(field, "slug", "Enter a valid slug.")
		}
	}
}

func UUID(field string) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		if !uuidPattern.MatchString(fmt.Sprint(values[field])) {
			errs.Add(field, "uuid", "Enter a valid UUID.")
		}
	}
}

func Choice(field string, choices ...any) Validator {
	return func(_ context.Context, values map[string]any, errs *Errors) {
		for _, choice := range choices {
			if reflect.DeepEqual(values[field], choice) {
				return
			}
		}
		errs.Add(field, "choice", "Select a valid choice.")
	}
}

func Unique(field string, check func(context.Context, any) (bool, error)) Validator {
	return func(ctx context.Context, values map[string]any, errs *Errors) {
		ok, err := check(ctx, values[field])
		if err != nil || !ok {
			errs.Add(field, "unique", "This value must be unique.")
		}
	}
}

func UniqueForDate(field string, check func(context.Context, time.Time, any) (bool, error)) Validator {
	return uniqueForPeriod(field, "unique_for_date", check)
}

func UniqueForMonth(field string, check func(context.Context, time.Time, any) (bool, error)) Validator {
	return uniqueForPeriod(field, "unique_for_month", check)
}

func UniqueForYear(field string, check func(context.Context, time.Time, any) (bool, error)) Validator {
	return uniqueForPeriod(field, "unique_for_year", check)
}

func Constraint(field string, check func(context.Context, map[string]any) (bool, error)) Validator {
	return func(ctx context.Context, values map[string]any, errs *Errors) {
		ok, err := check(ctx, values)
		if err != nil || !ok {
			errs.Add(field, "constraint", "Constraint validation failed.")
		}
	}
}

func Custom(field, code string, check func(context.Context, any) error) Validator {
	return func(ctx context.Context, values map[string]any, errs *Errors) {
		if err := check(ctx, values[field]); err != nil {
			errs.Add(field, code, err.Error())
		}
	}
}

func FieldLevel(field string, check func(context.Context, any) error) Validator {
	return Custom(field, "field", check)
}

func ModelLevel(code string, check func(context.Context, map[string]any) error) Validator {
	return func(ctx context.Context, values map[string]any, errs *Errors) {
		if err := check(ctx, values); err != nil {
			errs.Add("__all__", code, err.Error())
		}
	}
}

func uniqueForPeriod(field, code string, check func(context.Context, time.Time, any) (bool, error)) Validator {
	return func(ctx context.Context, values map[string]any, errs *Errors) {
		value, ok := values[field].(time.Time)
		if !ok {
			errs.Add(field, code, "Enter a valid date.")
			return
		}
		passed, err := check(ctx, value, values[field])
		if err != nil || !passed {
			errs.Add(field, code, "This value must be unique for the selected period.")
		}
	}
}

func numeric(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	default:
		return 0, false
	}
}

func lengthOf(value any) (int, bool) {
	switch typed := value.(type) {
	case string:
		return len(typed), true
	default:
		reflected := reflect.ValueOf(value)
		switch reflected.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return reflected.Len(), true
		default:
			return 0, false
		}
	}
}

func normalizeCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
