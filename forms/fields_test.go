package forms

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestFormFieldsCleanEverySupportedType(t *testing.T) {
	tests := []struct {
		name  string
		field *Field
		input any
		check func(any) bool
	}{
		{name: "boolean", field: BooleanField(FieldOptions{}), input: "true", check: func(v any) bool { return v == true }},
		{name: "char", field: CharField(FieldOptions{}), input: 42, check: func(v any) bool { return v == "42" }},
		{name: "choice", field: ChoiceField(FieldOptions{}, []Choice{{Value: "draft", Label: "Draft"}}), input: "draft", check: func(v any) bool { return v == "draft" }},
		{name: "typed choice", field: TypedChoiceField(FieldOptions{}, []Choice{{Value: "7", Label: "Seven"}}, func(v string) (any, error) { return "typed:" + v, nil }), input: "7", check: func(v any) bool { return v == "typed:7" }},
		{name: "multiple choice", field: MultipleChoiceField(FieldOptions{}, []Choice{{Value: "go"}, {Value: "api"}}), input: []any{"go", "api"}, check: func(v any) bool { return reflect.DeepEqual(v, []string{"go", "api"}) }},
		{name: "date", field: DateField(FieldOptions{}), input: "2026-06-28", check: func(v any) bool { return v.(time.Time).Format("2006-01-02") == "2026-06-28" }},
		{name: "datetime", field: DateTimeField(FieldOptions{}), input: "2026-06-28T10:00:00Z", check: func(v any) bool { return v.(time.Time).UTC().Hour() == 10 }},
		{name: "time", field: TimeField(FieldOptions{}), input: "10:30:00", check: func(v any) bool { return v.(time.Time).Format("15:04:05") == "10:30:00" }},
		{name: "duration", field: DurationField(FieldOptions{}), input: "2h", check: func(v any) bool { return v == 2*time.Hour }},
		{name: "decimal", field: DecimalField(FieldOptions{}), input: "10.25", check: func(v any) bool { return v == "10.25" }},
		{name: "email", field: EmailField(FieldOptions{}), input: "dev@example.com", check: func(v any) bool { return v == "dev@example.com" }},
		{name: "file", field: FileField(FieldOptions{}), input: UploadedFile{Name: "doc.txt", Content: []byte("x")}, check: func(v any) bool { return v.(UploadedFile).Name == "doc.txt" }},
		{name: "image", field: ImageField(FieldOptions{}), input: UploadedFile{Name: "avatar.png", Content: []byte{0x89, 'P', 'N', 'G'}}, check: func(v any) bool { return v.(UploadedFile).Name == "avatar.png" }},
		{name: "float", field: FloatField(FieldOptions{}), input: "3.5", check: func(v any) bool { return v == 3.5 }},
		{name: "integer", field: IntegerField(FieldOptions{}), input: "42", check: func(v any) bool { return v == int64(42) }},
		{name: "ip", field: GenericIPAddressField(FieldOptions{}), input: "192.0.2.1", check: func(v any) bool { return v == "192.0.2.1" }},
		{name: "json", field: JSONField(FieldOptions{}), input: `{"ok":true}`, check: func(v any) bool { return v.(map[string]any)["ok"] == true }},
		{name: "combo", field: ComboField(FieldOptions{}, CharField(FieldOptions{}), RegexField(FieldOptions{}, regexp.MustCompile(`^go$`))), input: "go", check: func(v any) bool { return v == "go" }},
		{name: "multi value", field: MultiValueField(FieldOptions{}, CharField(FieldOptions{}), IntegerField(FieldOptions{})), input: []any{"go", "7"}, check: func(v any) bool { return reflect.DeepEqual(v, []any{"go", int64(7)}) }},
		{name: "split datetime", field: SplitDateTimeField(FieldOptions{}), input: []any{"2026-06-28", "10:30:00"}, check: func(v any) bool { return v.(time.Time).Format(time.RFC3339) == "2026-06-28T10:30:00Z" }},
		{name: "model choice", field: ModelChoiceField(FieldOptions{}, []Choice{{Value: "1"}}), input: "1", check: func(v any) bool { return v == "1" }},
		{name: "model multiple choice", field: ModelMultipleChoiceField(FieldOptions{}, []Choice{{Value: "1"}, {Value: "2"}}), input: []string{"1", "2"}, check: func(v any) bool { return reflect.DeepEqual(v, []string{"1", "2"}) }},
		{name: "multiple file", field: MultipleFileField(FieldOptions{}), input: []UploadedFile{{Name: "a.txt"}, {Name: "b.txt"}}, check: func(v any) bool { return len(v.([]UploadedFile)) == 2 }},
		{name: "regex", field: RegexField(FieldOptions{}, regexp.MustCompile(`^go$`)), input: "go", check: func(v any) bool { return v == "go" }},
		{name: "slug", field: SlugField(FieldOptions{}), input: "go-api", check: func(v any) bool { return v == "go-api" }},
		{name: "url", field: URLField(FieldOptions{}), input: "https://example.com", check: func(v any) bool { return v == "https://example.com" }},
		{name: "uuid", field: UUIDField(FieldOptions{}), input: "550e8400-e29b-41d4-a716-446655440000", check: func(v any) bool { return v == "550e8400-e29b-41d4-a716-446655440000" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cleaned, err := test.field.Clean(test.input)
			if err != nil {
				t.Fatalf("Clean() error = %v", err)
			}
			if !test.check(cleaned) {
				t.Fatalf("cleaned = %#v", cleaned)
			}
		})
	}
}

func TestFormFieldValidationDisabledAndEmptyValues(t *testing.T) {
	required := CharField(FieldOptions{Required: true})
	if _, err := required.Clean(""); !errors.Is(err, ErrValidation) {
		t.Fatalf("required Clean() error = %v, want ErrValidation", err)
	}

	optional := CharField(FieldOptions{})
	cleaned, err := optional.Clean("")
	if err != nil || cleaned != nil {
		t.Fatalf("optional Clean() = %#v, %v; want nil", cleaned, err)
	}

	disabled := IntegerField(FieldOptions{Disabled: true, Initial: int64(9)})
	cleaned, err = disabled.Clean("100")
	if err != nil || cleaned != int64(9) {
		t.Fatalf("disabled Clean() = %#v, %v; want initial", cleaned, err)
	}

	validated := CharField(FieldOptions{Validators: []Validator{func(any) error { return errors.New("bad") }}})
	if _, err := validated.Clean("value"); !errors.Is(err, ErrValidation) {
		t.Fatalf("validator Clean() error = %v, want ErrValidation", err)
	}
}
