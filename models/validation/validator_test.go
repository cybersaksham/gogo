package validation

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"
)

func TestFieldValidatorsReturnStructuredErrors(t *testing.T) {
	values := map[string]any{
		"required": "",
		"type":     "not-int",
		"min":      1,
		"max":      10,
		"min_len":  "a",
		"max_len":  "abcd",
		"regex":    "abc",
		"email":    "bad",
		"url":      "bad",
		"slug":     "bad slug",
		"uuid":     "bad",
		"choice":   "z",
	}

	err := Validate(context.Background(), values,
		Required("required"),
		Type("type", 0),
		MinValue("min", 2),
		MaxValue("max", 9),
		MinLength("min_len", 2),
		MaxLength("max_len", 3),
		Regex("regex", regexp.MustCompile(`^\d+$`)),
		Email("email"),
		URL("url"),
		Slug("slug"),
		UUID("uuid"),
		Choice("choice", "a", "b"),
	)
	if err == nil {
		t.Fatalf("Validate() error = nil, want structured errors")
	}
	validationErrors := err.(*Errors)
	for _, field := range []string{"required", "type", "min", "max", "min_len", "max_len", "regex", "email", "url", "slug", "uuid", "choice"} {
		if !validationErrors.Has(field) {
			t.Fatalf("missing validation error for %s in %#v", field, validationErrors.Fields)
		}
	}
}

func TestUniqueDateConstraintCustomFieldAndModelValidators(t *testing.T) {
	values := map[string]any{
		"name":       "taken",
		"published":  time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC),
		"constraint": "bad",
		"custom":     "bad",
	}

	err := Validate(context.Background(), values,
		Unique("name", func(context.Context, any) (bool, error) { return false, nil }),
		UniqueForDate("published", func(context.Context, time.Time, any) (bool, error) { return false, nil }),
		UniqueForMonth("published", func(context.Context, time.Time, any) (bool, error) { return false, nil }),
		UniqueForYear("published", func(context.Context, time.Time, any) (bool, error) { return false, nil }),
		Constraint("constraint", func(context.Context, map[string]any) (bool, error) { return false, nil }),
		Custom("custom", "custom", func(context.Context, any) error { return errors.New("bad custom") }),
		FieldLevel("custom", func(context.Context, any) error { return errors.New("bad field") }),
		ModelLevel("model", func(context.Context, map[string]any) error { return errors.New("bad model") }),
	)
	if err == nil {
		t.Fatalf("Validate() error = nil, want structured errors")
	}
	validationErrors := err.(*Errors)
	for _, field := range []string{"name", "published", "constraint", "custom", "__all__"} {
		if !validationErrors.Has(field) {
			t.Fatalf("missing validation error for %s in %#v", field, validationErrors.Fields)
		}
	}
}

func TestValidateReturnsNilWhenAllValidatorsPass(t *testing.T) {
	err := Validate(context.Background(), map[string]any{"name": "ok"},
		Required("name"),
		Choice("name", "ok"),
	)
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
