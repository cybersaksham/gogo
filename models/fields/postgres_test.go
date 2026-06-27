package fields

import (
	"errors"
	"testing"
	"time"
)

func TestPostgresArrayAndHStoreFields(t *testing.T) {
	array := NewArrayField(Options{Name: "tags"}, NewTextField(Options{Name: "tag"}))
	if got := array.ColumnType("postgres"); got != "text[]" {
		t.Fatalf("array ColumnType(postgres) = %q, want text[]", got)
	}
	if err := array.RequireDialect("sqlite"); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("RequireDialect(sqlite) error = %v, want ErrInvalidField", err)
	}
	if err := array.Validate([]string{"go", "django"}); err != nil {
		t.Fatalf("array Validate() error = %v", err)
	}
	if err := array.Validate("not-array"); !errors.Is(err, ErrValidation) {
		t.Fatalf("array Validate(string) error = %v, want ErrValidation", err)
	}

	hstore := NewHStoreField(Options{Name: "labels"})
	if got := hstore.ColumnType("postgres"); got != "hstore" {
		t.Fatalf("hstore ColumnType(postgres) = %q, want hstore", got)
	}
	if err := hstore.Validate(map[string]string{"env": "prod"}); err != nil {
		t.Fatalf("hstore Validate() error = %v", err)
	}
}

func TestPostgresRangeFields(t *testing.T) {
	tests := []struct {
		name       string
		field      *RangeField
		columnType string
		value      Range
	}{
		{"integer", NewIntegerRangeField(Options{Name: "r"}), "int4range", Range{Lower: int64(1), Upper: int64(10), Bounds: "[)"}},
		{"big integer", NewBigIntegerRangeField(Options{Name: "r"}), "int8range", Range{Lower: int64(1), Upper: int64(10), Bounds: "[)"}},
		{"decimal", NewDecimalRangeField(Options{Name: "r"}), "numrange", Range{Lower: "1.00", Upper: "2.00", Bounds: "[)"}},
		{"date", NewDateRangeField(Options{Name: "r"}), "daterange", Range{Lower: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Upper: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Bounds: "[)"}},
		{"datetime", NewDateTimeRangeField(Options{Name: "r"}), "tstzrange", Range{Lower: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Upper: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Bounds: "[)"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.field.ColumnType("postgres"); got != test.columnType {
				t.Fatalf("ColumnType(postgres) = %q, want %q", got, test.columnType)
			}
			if err := test.field.Validate(test.value); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if err := test.field.RequireDialect("mysql"); !errors.Is(err, ErrInvalidField) {
				t.Fatalf("RequireDialect(mysql) error = %v, want ErrInvalidField", err)
			}
		})
	}
}

func TestPostgresRangeRejectsInvalidBounds(t *testing.T) {
	field := NewIntegerRangeField(Options{Name: "r"})
	err := field.Validate(Range{Lower: int64(10), Upper: int64(1), Bounds: "[)"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(invalid range) error = %v, want ErrValidation", err)
	}
}
