package fields

import (
	"errors"
	"testing"
)

func TestBooleanFieldValidationAndConversion(t *testing.T) {
	field := NewBooleanField(Options{Name: "active"})

	if err := field.Validate(true); err != nil {
		t.Fatalf("Validate(true) error = %v", err)
	}
	if err := field.Validate("true"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(string) error = %v, want ErrValidation", err)
	}
	value, err := field.ToDB(true)
	if err != nil || value != true {
		t.Fatalf("ToDB(true) = (%#v, %v), want true nil", value, err)
	}
}

func TestCharFieldEnforcesMaxLength(t *testing.T) {
	field := NewCharField(Options{Name: "title"}, 3)

	if err := field.Validate("abc"); err != nil {
		t.Fatalf("Validate(abc) error = %v", err)
	}
	if err := field.Validate("abcd"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(abcd) error = %v, want ErrValidation", err)
	}
	if got := field.ColumnType("postgres"); got != "varchar(3)" {
		t.Fatalf("ColumnType(postgres) = %q, want varchar(3)", got)
	}
}

func TestTextFieldSupportsChoicesAndBlankValues(t *testing.T) {
	choices, err := NewChoices(TextChoice("draft", "Draft"))
	if err != nil {
		t.Fatalf("NewChoices() error = %v", err)
	}
	field := NewTextField(Options{Name: "status", Blank: true, Choices: choices.All()})

	if err := field.Validate(""); err != nil {
		t.Fatalf("Validate(blank) error = %v", err)
	}
	if err := field.Validate("draft"); err != nil {
		t.Fatalf("Validate(draft) error = %v", err)
	}
	if err := field.Validate("published"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(published) error = %v, want ErrValidation", err)
	}
}

func TestFormatFieldsValidateValues(t *testing.T) {
	tests := []struct {
		name    string
		field   Field
		valid   string
		invalid string
	}{
		{"email", NewEmailField(Options{Name: "email"}, 254), "user@example.com", "not-email"},
		{"url", NewURLField(Options{Name: "site"}, 200), "https://example.com/path", "://bad"},
		{"slug", NewSlugField(Options{Name: "slug"}, 50), "post-1_slug", "bad slug"},
		{"uuid", NewUUIDField(Options{Name: "uuid"}), "550e8400-e29b-41d4-a716-446655440000", "bad-uuid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.field.Validate(test.valid); err != nil {
				t.Fatalf("Validate(valid) error = %v", err)
			}
			if err := test.field.Validate(test.invalid); !errors.Is(err, ErrValidation) {
				t.Fatalf("Validate(invalid) error = %v, want ErrValidation", err)
			}
			value, err := test.field.ToDB(test.valid)
			if err != nil || value != test.valid {
				t.Fatalf("ToDB(valid) = (%#v, %v), want original string", value, err)
			}
		})
	}
}
