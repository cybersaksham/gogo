package fields

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
)

var (
	slugPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// BooleanField stores boolean values.
type BooleanField struct {
	*BaseField
}

// NewBooleanField creates a boolean field.
func NewBooleanField(options Options) *BooleanField {
	return &BooleanField{BaseField: NewBaseField("boolean", options, map[string]string{"postgres": "boolean", "sqlite": "boolean"})}
}

func (f *BooleanField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || value == nil && f.options.Null {
		return err
	}
	if _, ok := value.(bool); !ok {
		return fmt.Errorf("%w: %s must be boolean", ErrValidation, f.Name())
	}
	return nil
}

func (f *BooleanField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return value.(bool), nil
}

func (f *BooleanField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *BooleanField) Clone() Field {
	return &BooleanField{BaseField: f.BaseField.Clone().(*BaseField)}
}

type stringField struct {
	*BaseField
	maxLength int
	validate  func(string) error
}

// NewCharField creates a bounded string field.
func NewCharField(options Options, maxLength int) Field {
	return newStringField("char", options, maxLength, map[string]string{"postgres": fmt.Sprintf("varchar(%d)", maxLength), "sqlite": "text"}, nil)
}

// NewTextField creates an unbounded text field.
func NewTextField(options Options) Field {
	return newStringField("text", options, 0, map[string]string{"postgres": "text", "sqlite": "text"}, nil)
}

// NewEmailField creates an email field.
func NewEmailField(options Options, maxLength int) Field {
	return newStringField("email", options, maxLength, map[string]string{"postgres": fmt.Sprintf("varchar(%d)", maxLength), "sqlite": "text"}, validateEmail)
}

// NewURLField creates a URL field.
func NewURLField(options Options, maxLength int) Field {
	return newStringField("url", options, maxLength, map[string]string{"postgres": fmt.Sprintf("varchar(%d)", maxLength), "sqlite": "text"}, validateURL)
}

// NewSlugField creates a slug field.
func NewSlugField(options Options, maxLength int) Field {
	return newStringField("slug", options, maxLength, map[string]string{"postgres": fmt.Sprintf("varchar(%d)", maxLength), "sqlite": "text"}, validateSlug)
}

// NewUUIDField creates a UUID field.
func NewUUIDField(options Options) Field {
	return newStringField("uuid", options, 36, map[string]string{"postgres": "uuid", "sqlite": "text"}, validateUUID)
}

func newStringField(kind string, options Options, maxLength int, columnTypes map[string]string, validator func(string) error) Field {
	return &stringField{BaseField: NewBaseField(kind, options, columnTypes), maxLength: maxLength, validate: validator}
}

func (f *stringField) Validate(value any) error {
	if err := f.BaseField.Validate(value); err != nil || f.emptyAllowed(value) {
		return err
	}
	stringValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("%w: %s must be text", ErrValidation, f.Name())
	}
	if f.maxLength > 0 && len(stringValue) > f.maxLength {
		return fmt.Errorf("%w: %s exceeds max length", ErrValidation, f.Name())
	}
	if f.validate != nil {
		if err := f.validate(stringValue); err != nil {
			return fmt.Errorf("%w: %v", ErrValidation, err)
		}
	}
	return nil
}

func (f *stringField) ToDB(value any) (any, error) {
	if err := f.Validate(value); err != nil {
		return nil, err
	}
	return value.(string), nil
}

func (f *stringField) FromDB(value any) (any, error) {
	return f.ToDB(value)
}

func (f *stringField) Clone() Field {
	return &stringField{
		BaseField: f.BaseField.Clone().(*BaseField),
		maxLength: f.maxLength,
		validate:  f.validate,
	}
}

func (f *stringField) emptyAllowed(value any) bool {
	return value == nil && f.options.Null || value == "" && f.options.Blank
}

func validateEmail(value string) error {
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return fmt.Errorf("invalid email")
	}
	return nil
}

func validateURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL")
	}
	return nil
}

func validateSlug(value string) error {
	if !slugPattern.MatchString(value) {
		return fmt.Errorf("invalid slug")
	}
	return nil
}

func validateUUID(value string) error {
	if !uuidPattern.MatchString(value) {
		return fmt.Errorf("invalid UUID")
	}
	return nil
}
