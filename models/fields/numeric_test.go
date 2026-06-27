package fields

import (
	"errors"
	"testing"
)

func TestAutoAndIntegerFieldsColumnTypesAndValidation(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		valid    any
		invalid  any
		postgres string
		sqlite   string
	}{
		{"auto", NewAutoField(Options{Name: "id"}), int64(1), int64(-1), "serial", "integer"},
		{"big auto", NewBigAutoField(Options{Name: "id"}), int64(1), int64(-1), "bigserial", "integer"},
		{"small auto", NewSmallAutoField(Options{Name: "id"}), int64(1), int64(-1), "smallserial", "integer"},
		{"integer", NewIntegerField(Options{Name: "count"}), int64(2147483647), int64(2147483648), "integer", "integer"},
		{"big integer", NewBigIntegerField(Options{Name: "count"}), int64(9223372036854775807), "not-int", "bigint", "integer"},
		{"small integer", NewSmallIntegerField(Options{Name: "count"}), int64(32767), int64(32768), "smallint", "integer"},
		{"positive integer", NewPositiveIntegerField(Options{Name: "count"}), int64(0), int64(-1), "integer", "integer"},
		{"positive big integer", NewPositiveBigIntegerField(Options{Name: "count"}), int64(0), int64(-1), "bigint", "integer"},
		{"positive small integer", NewPositiveSmallIntegerField(Options{Name: "count"}), int64(0), int64(-1), "smallint", "integer"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.field.ColumnType("postgres"); got != test.postgres {
				t.Fatalf("ColumnType(postgres) = %q, want %q", got, test.postgres)
			}
			if got := test.field.ColumnType("sqlite"); got != test.sqlite {
				t.Fatalf("ColumnType(sqlite) = %q, want %q", got, test.sqlite)
			}
			if err := test.field.Validate(test.valid); err != nil {
				t.Fatalf("Validate(valid) error = %v", err)
			}
			if err := test.field.Validate(test.invalid); !errors.Is(err, ErrValidation) {
				t.Fatalf("Validate(invalid) error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestIntegerFieldConvertsToDatabaseInteger(t *testing.T) {
	field := NewIntegerField(Options{Name: "count"})

	value, err := field.ToDB(int(12))
	if err != nil {
		t.Fatalf("ToDB() error = %v", err)
	}
	if value != int64(12) {
		t.Fatalf("ToDB() = %#v, want int64(12)", value)
	}
}

func TestDecimalFieldValidatesDigitsAndPlaces(t *testing.T) {
	field := NewDecimalField(Options{Name: "price"}, 5, 2)

	if err := field.Validate("123.45"); err != nil {
		t.Fatalf("Validate(123.45) error = %v", err)
	}
	if err := field.Validate("1234.56"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(1234.56) error = %v, want ErrValidation", err)
	}
	if err := field.Validate("1.234"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(1.234) error = %v, want ErrValidation", err)
	}
	if got := field.ColumnType("postgres"); got != "numeric(5,2)" {
		t.Fatalf("ColumnType(postgres) = %q, want numeric(5,2)", got)
	}
}

func TestFloatFieldConvertsToDatabaseFloat(t *testing.T) {
	field := NewFloatField(Options{Name: "score"})

	value, err := field.ToDB(12)
	if err != nil {
		t.Fatalf("ToDB() error = %v", err)
	}
	if value != float64(12) {
		t.Fatalf("ToDB() = %#v, want float64(12)", value)
	}
	if err := field.Validate("not-float"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(not-float) error = %v, want ErrValidation", err)
	}
}
