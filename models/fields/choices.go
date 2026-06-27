package fields

import (
	"fmt"
	"reflect"
)

// Choices is an immutable collection of field choices.
type Choices struct {
	values []Choice
}

// NewChoices validates and returns a choice collection.
func NewChoices(values ...Choice) (Choices, error) {
	choices := Choices{values: append([]Choice(nil), values...)}
	if err := ValidateChoices(choices.values); err != nil {
		return Choices{}, err
	}
	return choices, nil
}

// TextChoice creates a string choice.
func TextChoice(value, label string) Choice {
	return Choice{Value: value, Label: label}
}

// IntegerChoice creates an integer choice.
func IntegerChoice(value int, label string) Choice {
	return Choice{Value: value, Label: label}
}

// BlankChoice creates a blank string choice.
func BlankChoice(label string) Choice {
	return Choice{Value: "", Label: label}
}

// EnumChoice creates a choice for a custom enumeration value.
func EnumChoice(value any, label string) Choice {
	return Choice{Value: value, Label: label}
}

// Group marks choices with a group label.
func Group(label string, values ...Choice) []Choice {
	grouped := append([]Choice(nil), values...)
	for i := range grouped {
		grouped[i].Group = label
	}
	return grouped
}

// Append returns a new collection with extra choices.
func (c Choices) Append(values ...Choice) (Choices, error) {
	next := append(c.All(), values...)
	return NewChoices(next...)
}

// All returns copied choices.
func (c Choices) All() []Choice {
	return append([]Choice(nil), c.values...)
}

// Label returns the display label for a choice value.
func (c Choices) Label(value any) (string, bool) {
	return ChoiceLabel(c.values, value)
}

// ChoiceLabel returns the display label for a choice value.
func ChoiceLabel(values []Choice, value any) (string, bool) {
	for _, choice := range values {
		if reflect.DeepEqual(choice.Value, value) {
			return choice.Label, true
		}
	}
	return "", false
}

// ValidateChoices rejects duplicate choice values.
func ValidateChoices(values []Choice) error {
	for i, left := range values {
		for j := i + 1; j < len(values); j++ {
			if reflect.DeepEqual(left.Value, values[j].Value) {
				return fmt.Errorf("%w: duplicate choice value %v", ErrInvalidField, left.Value)
			}
		}
	}
	return nil
}

// ValidateChoiceDefault validates that a default belongs to configured choices.
func ValidateChoiceDefault(options Options) error {
	if len(options.Choices) == 0 || options.Default == nil {
		return nil
	}
	if _, ok := ChoiceLabel(options.Choices, options.Default); !ok {
		return fmt.Errorf("%w: default %v is not a valid choice for %s", ErrInvalidField, options.Default, options.Name)
	}
	return nil
}
