package fields

import (
	"errors"
	"testing"
)

type status string

const archived status = "archived"

func TestSimpleTextChoicesReturnDisplayLabels(t *testing.T) {
	choices, err := NewChoices(
		TextChoice("draft", "Draft"),
		TextChoice("published", "Published"),
	)
	if err != nil {
		t.Fatalf("NewChoices() error = %v", err)
	}

	label, ok := choices.Label("published")
	if !ok || label != "Published" {
		t.Fatalf("Label(published) = (%q, %v), want Published true", label, ok)
	}
}

func TestGroupedIntegerBlankAndEnumChoices(t *testing.T) {
	choices, err := NewChoices(
		Group("Numbers", IntegerChoice(1, "One"), IntegerChoice(2, "Two"))...,
	)
	if err != nil {
		t.Fatalf("NewChoices(grouped) error = %v", err)
	}
	choices, err = choices.Append(BlankChoice("---------"), EnumChoice(archived, "Archived"))
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	if choices.All()[0].Group != "Numbers" {
		t.Fatalf("Group = %q, want Numbers", choices.All()[0].Group)
	}
	if label, ok := choices.Label(1); !ok || label != "One" {
		t.Fatalf("Label(1) = (%q, %v), want One true", label, ok)
	}
	if label, ok := choices.Label(""); !ok || label != "---------" {
		t.Fatalf("Label(blank) = (%q, %v), want blank label", label, ok)
	}
	if label, ok := choices.Label(archived); !ok || label != "Archived" {
		t.Fatalf("Label(enum) = (%q, %v), want Archived", label, ok)
	}
}

func TestChoicesRejectDuplicateValues(t *testing.T) {
	_, err := NewChoices(TextChoice("draft", "Draft"), TextChoice("draft", "Draft again"))
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("NewChoices() error = %v, want ErrInvalidField", err)
	}
}

func TestValidateChoiceDefaultRejectsInvalidDefault(t *testing.T) {
	choices, err := NewChoices(TextChoice("draft", "Draft"))
	if err != nil {
		t.Fatalf("NewChoices() error = %v", err)
	}
	err = ValidateChoiceDefault(Options{Name: "status", Default: "published", Choices: choices.All()})
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("ValidateChoiceDefault() error = %v, want ErrInvalidField", err)
	}
}

func TestBaseFieldUsesChoicesForValidation(t *testing.T) {
	choices, err := NewChoices(TextChoice("draft", "Draft"))
	if err != nil {
		t.Fatalf("NewChoices() error = %v", err)
	}
	field := NewBaseField("text", Options{Name: "status", Choices: choices.All()}, nil)

	if err := field.Validate("draft"); err != nil {
		t.Fatalf("Validate(draft) error = %v", err)
	}
	if err := field.Validate("published"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(published) error = %v, want ErrValidation", err)
	}
}
