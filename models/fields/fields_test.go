package fields

import (
	"errors"
	"testing"
)

func TestBaseFieldAppliesOptionDefaults(t *testing.T) {
	field := NewBaseField("text", Options{Name: "title"}, map[string]string{"postgres": "text"})

	if field.Name() != "title" {
		t.Fatalf("Name() = %q, want title", field.Name())
	}
	if field.ColumnName() != "title" {
		t.Fatalf("ColumnName() = %q, want title", field.ColumnName())
	}
	if !field.IsEditable() {
		t.Fatalf("IsEditable() = false, want true")
	}
	if !field.IsSerializable() {
		t.Fatalf("IsSerializable() = false, want true")
	}
	if got := field.ColumnType("postgres"); got != "text" {
		t.Fatalf("ColumnType(postgres) = %q, want text", got)
	}
}

func TestBaseFieldUsesExplicitColumnName(t *testing.T) {
	field := NewBaseField("integer", Options{Name: "userID", Column: "user_id"}, map[string]string{"default": "integer"})

	if field.ColumnName() != "user_id" {
		t.Fatalf("ColumnName() = %q, want user_id", field.ColumnName())
	}
}

func TestBaseFieldValidationRejectsNullBlankAndValidatorErrors(t *testing.T) {
	required := NewBaseField("text", Options{Name: "title"}, nil)
	if err := required.Validate(nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(nil) error = %v, want ErrValidation", err)
	}
	if err := required.Validate(""); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(blank) error = %v, want ErrValidation", err)
	}

	nullable := NewBaseField("text", Options{Name: "title", Null: true, Blank: true}, nil)
	if err := nullable.Validate(nil); err != nil {
		t.Fatalf("nullable Validate(nil) error = %v", err)
	}
	if err := nullable.Validate(""); err != nil {
		t.Fatalf("blank Validate(\"\") error = %v", err)
	}

	invalid := NewBaseField("text", Options{
		Name: "title",
		Validators: []Validator{
			func(any) error { return errors.New("not valid") },
		},
	}, nil)
	if err := invalid.Validate("value"); !errors.Is(err, ErrValidation) {
		t.Fatalf("validator error = %v, want ErrValidation", err)
	}
}

func TestBaseFieldCloneIsIndependent(t *testing.T) {
	editable := false
	field := NewBaseField("text", Options{
		Name:          "title",
		Editable:      &editable,
		ErrorMessages: map[string]string{"required": "Required"},
	}, nil)

	clone := field.Clone().(*BaseField)
	clone.options.Name = "changed"
	clone.options.ErrorMessages["required"] = "Changed"

	if field.Name() != "title" {
		t.Fatalf("original Name() = %q, want title", field.Name())
	}
	if field.Options().ErrorMessages["required"] != "Required" {
		t.Fatalf("original error message changed: %#v", field.Options().ErrorMessages)
	}
	if field.IsEditable() {
		t.Fatalf("IsEditable() = true, want explicit false")
	}
}
