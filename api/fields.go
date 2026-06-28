package api

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FieldValidator func(any) error

// FieldOptions configures serializer field behavior.
type FieldOptions struct {
	Required      bool
	AllowNull     bool
	AllowBlank    bool
	Default       any
	Source        string
	Label         string
	HelpText      string
	ReadOnly      bool
	WriteOnly     bool
	Validators    []FieldValidator
	ErrorMessages map[string]string
}

// SerializerField parses and renders one serializer value.
type SerializerField struct {
	Name       string
	Kind       string
	Options    FieldOptions
	Choices    []string
	Child      *SerializerField
	Nested     *Serializer
	MethodFunc func(map[string]any) any
}

func BooleanField(name string, options FieldOptions) SerializerField {
	return field(name, "boolean", options)
}
func IntegerField(name string, options FieldOptions) SerializerField {
	return field(name, "integer", options)
}
func FloatField(name string, options FieldOptions) SerializerField {
	return field(name, "float", options)
}
func DecimalField(name string, options FieldOptions) SerializerField {
	return field(name, "decimal", options)
}
func StringField(name string, options FieldOptions) SerializerField {
	return field(name, "string", options)
}
func EmailField(name string, options FieldOptions) SerializerField {
	return field(name, "email", options)
}
func URLField(name string, options FieldOptions) SerializerField { return field(name, "url", options) }
func SlugField(name string, options FieldOptions) SerializerField {
	return field(name, "slug", options)
}
func UUIDField(name string, options FieldOptions) SerializerField {
	return field(name, "uuid", options)
}
func DateField(name string, options FieldOptions) SerializerField {
	return field(name, "date", options)
}
func DateTimeField(name string, options FieldOptions) SerializerField {
	return field(name, "datetime", options)
}
func TimeField(name string, options FieldOptions) SerializerField {
	return field(name, "time", options)
}
func DurationField(name string, options FieldOptions) SerializerField {
	return field(name, "duration", options)
}
func JSONField(name string, options FieldOptions) SerializerField {
	return field(name, "json", options)
}
func DictField(name string, options FieldOptions) SerializerField {
	return field(name, "dict", options)
}
func PrimaryKeyRelatedField(name string, options FieldOptions) SerializerField {
	return field(name, "primary_key_related", options)
}
func SlugRelatedField(name string, options FieldOptions) SerializerField {
	return field(name, "slug_related", options)
}
func HyperlinkedRelatedField(name string, options FieldOptions) SerializerField {
	return field(name, "hyperlinked_related", options)
}
func FileField(name string, options FieldOptions) SerializerField {
	return field(name, "file", options)
}
func ImageField(name string, options FieldOptions) SerializerField {
	return field(name, "image", options)
}

func ChoiceField(name string, options FieldOptions, choices []string) SerializerField {
	f := field(name, "choice", options)
	f.Choices = append([]string(nil), choices...)
	return f
}

func MultipleChoiceField(name string, options FieldOptions, choices []string) SerializerField {
	f := field(name, "multiple_choice", options)
	f.Choices = append([]string(nil), choices...)
	return f
}

func ListField(name string, options FieldOptions, child SerializerField) SerializerField {
	f := field(name, "list", options)
	f.Child = &child
	return f
}

func NestedObjectField(name string, options FieldOptions, serializer *Serializer) SerializerField {
	f := field(name, "nested", options)
	f.Nested = serializer
	return f
}

func MethodField(name string, method func(map[string]any) any) SerializerField {
	return SerializerField{Name: name, Kind: "method", Options: FieldOptions{ReadOnly: true}, MethodFunc: method}
}

func field(name, kind string, options FieldOptions) SerializerField {
	return SerializerField{Name: name, Kind: kind, Options: options}
}

func (f SerializerField) source() string {
	if f.Options.Source != "" {
		return f.Options.Source
	}
	return f.Name
}

func (f SerializerField) parse(value any) (any, []string) {
	if value == nil {
		if f.Options.AllowNull {
			return nil, nil
		}
		return nil, []string{"null"}
	}
	var parsed any
	var err error
	switch f.Kind {
	case "boolean":
		parsed, err = parseBool(value)
	case "integer", "primary_key_related":
		parsed, err = parseInt(value)
	case "float":
		parsed, err = parseFloat(value)
	case "decimal":
		parsed, err = parseDecimal(value)
	case "string":
		parsed, err = parseString(value, f.Options.AllowBlank)
	case "email":
		parsed, err = parseEmail(value)
	case "url", "hyperlinked_related":
		parsed, err = parseURL(value)
	case "slug", "slug_related":
		parsed, err = parseSlug(value)
	case "uuid":
		parsed, err = parseUUID(value)
	case "date":
		parsed, err = parseDate(value)
	case "datetime":
		parsed, err = parseDateTime(value)
	case "time":
		parsed, err = parseClock(value)
	case "duration":
		parsed, err = parseDuration(value)
	case "choice":
		parsed, err = parseChoice(value, f.Choices)
	case "multiple_choice":
		parsed, err = parseMultipleChoice(value, f.Choices)
	case "json":
		parsed = value
	case "list":
		parsed, err = parseList(value, f.Child)
	case "dict":
		parsed, err = parseDict(value)
	case "nested":
		parsed, err = parseNested(value, f.Nested)
	case "file":
		parsed, err = parseFile(value)
	case "image":
		parsed, err = parseImage(value)
	default:
		parsed = value
	}
	if err != nil {
		return nil, []string{err.Error()}
	}
	for _, validator := range f.Options.Validators {
		if err := validator(parsed); err != nil {
			return nil, []string{err.Error()}
		}
	}
	return parsed, nil
}

func (f SerializerField) render(obj map[string]any) any {
	if f.Kind == "method" && f.MethodFunc != nil {
		return f.MethodFunc(obj)
	}
	return obj[f.source()]
}

func parseBool(value any) (bool, error) {
	switch typed := value.(type) {
	case bool:
		return typed, nil
	case string:
		return strconv.ParseBool(typed)
	default:
		return false, fmt.Errorf("invalid boolean")
	}
}

func parseInt(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	case float64:
		return int64(typed), nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, fmt.Errorf("invalid integer")
	}
}

func parseFloat(value any) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case string:
		return strconv.ParseFloat(typed, 64)
	default:
		return 0, fmt.Errorf("invalid float")
	}
}

func parseDecimal(value any) (string, error) {
	text := fmt.Sprint(value)
	if _, err := strconv.ParseFloat(text, 64); err != nil {
		return "", fmt.Errorf("invalid decimal")
	}
	return text, nil
}

func parseString(value any, allowBlank bool) (string, error) {
	text := fmt.Sprint(value)
	if text == "" && !allowBlank {
		return "", fmt.Errorf("blank")
	}
	return text, nil
}

func parseEmail(value any) (string, error) {
	text, err := parseString(value, false)
	if err != nil {
		return "", err
	}
	if !strings.Contains(text, "@") {
		return "", fmt.Errorf("invalid email")
	}
	return text, nil
}

func parseURL(value any) (string, error) {
	text, err := parseString(value, false)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(text)
	if err != nil || parsed.String() == "" {
		return "", fmt.Errorf("invalid url")
	}
	return text, nil
}

var slugPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func parseSlug(value any) (string, error) {
	text, err := parseString(value, false)
	if err != nil {
		return "", err
	}
	if !slugPattern.MatchString(text) {
		return "", fmt.Errorf("invalid slug")
	}
	return text, nil
}

func parseUUID(value any) (string, error) {
	text, err := parseString(value, false)
	if err != nil {
		return "", err
	}
	if _, err := uuid.Parse(text); err != nil {
		return "", fmt.Errorf("invalid uuid")
	}
	return text, nil
}

func parseDate(value any) (time.Time, error) {
	return time.Parse("2006-01-02", fmt.Sprint(value))
}

func parseDateTime(value any) (time.Time, error) {
	return time.Parse(time.RFC3339, fmt.Sprint(value))
}

func parseClock(value any) (time.Time, error) {
	return time.Parse("15:04:05", fmt.Sprint(value))
}

func parseDuration(value any) (time.Duration, error) {
	return time.ParseDuration(fmt.Sprint(value))
}

func parseChoice(value any, choices []string) (string, error) {
	text := fmt.Sprint(value)
	for _, choice := range choices {
		if text == choice {
			return text, nil
		}
	}
	return "", fmt.Errorf("invalid choice")
}

func parseMultipleChoice(value any, choices []string) ([]string, error) {
	var values []string
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			values = append(values, fmt.Sprint(item))
		}
	case []string:
		values = append(values, typed...)
	default:
		return nil, fmt.Errorf("invalid multiple choice")
	}
	for _, value := range values {
		if _, err := parseChoice(value, choices); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func parseList(value any, child *SerializerField) ([]any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid list")
	}
	out := make([]any, len(items))
	for i, item := range items {
		parsed, errors := child.parse(item)
		if len(errors) > 0 {
			return nil, fmt.Errorf("invalid list item")
		}
		out[i] = parsed
	}
	return out, nil
}

func parseDict(value any) (map[string]any, error) {
	dict, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid dict")
	}
	return dict, nil
}

func parseNested(value any, serializer *Serializer) (map[string]any, error) {
	dict, ok := value.(map[string]any)
	if !ok || serializer == nil {
		return nil, fmt.Errorf("invalid object")
	}
	validated, errors, ok := serializer.Validate(dict)
	if !ok {
		return nil, nestedValidationError{errors: errors}
	}
	return validated, nil
}

func parseFile(value any) (UploadedFile, error) {
	file, ok := value.(UploadedFile)
	if !ok || file.Filename == "" {
		return UploadedFile{}, fmt.Errorf("invalid file")
	}
	return file, nil
}

func parseImage(value any) (UploadedFile, error) {
	file, err := parseFile(value)
	if err != nil {
		return UploadedFile{}, err
	}
	lower := strings.ToLower(file.Filename)
	if !(strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif")) {
		return UploadedFile{}, fmt.Errorf("invalid image")
	}
	return file, nil
}

type nestedValidationError struct {
	errors map[string][]string
}

func (e nestedValidationError) Error() string { return "nested validation" }
