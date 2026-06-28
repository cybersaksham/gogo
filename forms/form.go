package forms

import (
	"fmt"
	"html"
	"reflect"
	"sort"
	"strings"
)

// FormOptions configures a form instance.
type FormOptions struct {
	Fields     map[string]*Field
	FieldOrder []string
	Data       map[string]any
	Initial    map[string]any
	Prefix     string
	Bound      bool
	Groups     []FieldGroup
	Clean      FormCleanFunc
}

// FieldGroup defines a named group of form fields.
type FieldGroup struct {
	Name   string
	Fields []string
}

// BoundFieldGroup is a field group resolved to bound fields.
type BoundFieldGroup struct {
	Name   string
	Fields []BoundField
}

// Form binds data to fields and coordinates validation and rendering.
type Form struct {
	CleanedData map[string]any

	fields         map[string]*Field
	fieldOrder     []string
	data           map[string]any
	initial        map[string]any
	prefix         string
	bound          bool
	groups         []FieldGroup
	clean          FormCleanFunc
	errors         ErrorDict
	nonFieldErrors NonFieldErrorList
	cleaned        bool
}

func NewForm(options FormOptions) *Form {
	fields := cloneFieldMap(options.Fields)
	initial := cloneAnyMap(options.Initial)
	data := cloneAnyMap(options.Data)
	form := &Form{
		CleanedData: make(map[string]any),
		fields:      fields,
		fieldOrder:  resolveFieldOrder(fields, options.FieldOrder),
		data:        data,
		initial:     initial,
		prefix:      options.Prefix,
		bound:       options.Bound || options.Data != nil,
		groups:      cloneFieldGroups(options.Groups),
		clean:       options.Clean,
		errors:      ErrorDict{},
	}
	return form
}

func (f *Form) IsBound() bool {
	return f != nil && f.bound
}

func (f *Form) FieldNames() []string {
	if f == nil {
		return nil
	}
	return append([]string(nil), f.fieldOrder...)
}

func (f *Form) BoundField(name string) BoundField {
	if f == nil {
		return BoundField{Name: name}
	}
	return BoundField{Form: f, Name: name, Field: f.fields[name]}
}

func (f *Form) VisibleFields() []BoundField {
	return f.boundFieldsMatching(false)
}

func (f *Form) HiddenFields() []BoundField {
	return f.boundFieldsMatching(true)
}

func (f *Form) FieldGroups() []BoundFieldGroup {
	if f == nil {
		return nil
	}
	groups := make([]BoundFieldGroup, 0, len(f.groups))
	for _, group := range f.groups {
		boundGroup := BoundFieldGroup{Name: group.Name}
		for _, fieldName := range group.Fields {
			if _, ok := f.fields[fieldName]; ok {
				boundGroup.Fields = append(boundGroup.Fields, f.BoundField(fieldName))
			}
		}
		groups = append(groups, boundGroup)
	}
	return groups
}

func (f *Form) Media() Media {
	if f == nil {
		return Media{}
	}
	var media Media
	for _, name := range f.fieldOrder {
		widget := f.BoundField(name).widget()
		if provider, ok := widget.(MediaProvider); ok {
			media = media.Merge(provider.Media())
		}
	}
	return media
}

func (f *Form) IsValid() bool {
	if f == nil || !f.bound {
		return false
	}
	f.FullClean()
	return !f.errors.HasErrors() && len(f.nonFieldErrors) == 0
}

func (f *Form) FullClean() {
	if f == nil || f.cleaned {
		return
	}
	f.CleanedData = make(map[string]any)
	f.errors = ErrorDict{}
	f.nonFieldErrors = nil
	if !f.bound {
		f.cleaned = true
		return
	}
	for _, name := range f.fieldOrder {
		field := f.fields[name]
		if field == nil {
			continue
		}
		cleaned, err := field.Clean(f.dataValue(name))
		if err != nil {
			f.AddError(name, err)
			continue
		}
		f.CleanedData[name] = cleaned
	}
	if f.clean != nil {
		if err := f.clean(f); err != nil {
			f.AddNonFieldError(err)
		}
	}
	f.cleaned = true
}

func (f *Form) Errors() ErrorDict {
	if f == nil {
		return ErrorDict{}
	}
	f.FullClean()
	cloned := make(ErrorDict, len(f.errors))
	for field, errors := range f.errors {
		cloned[field] = append(ErrorList(nil), errors...)
	}
	return cloned
}

func (f *Form) FieldErrors(name string) ErrorList {
	if f == nil {
		return nil
	}
	f.FullClean()
	return f.errors.Get(name)
}

func (f *Form) NonFieldErrors() NonFieldErrorList {
	if f == nil {
		return nil
	}
	f.FullClean()
	return append(NonFieldErrorList(nil), f.nonFieldErrors...)
}

func (f *Form) AddError(field string, err error) {
	if f == nil || err == nil {
		return
	}
	if f.errors == nil {
		f.errors = ErrorDict{}
	}
	f.errors.Add(field, err)
}

func (f *Form) AddNonFieldError(err error) {
	if f == nil || err == nil {
		return
	}
	f.nonFieldErrors = append(f.nonFieldErrors, normalizeValidationError(err))
}

func (f *Form) ChangedData() []string {
	if f == nil || !f.bound {
		return nil
	}
	changed := make([]string, 0)
	for _, name := range f.fieldOrder {
		if valuesChanged(f.dataValue(name), f.initialValue(name)) {
			changed = append(changed, name)
		}
	}
	return changed
}

func (f *Form) Render() string {
	if f == nil {
		return ""
	}
	var builder strings.Builder
	for _, field := range f.VisibleFields() {
		builder.WriteString(`<div class="form-field">`)
		builder.WriteString(field.Errors().HTML())
		builder.WriteString(field.LabelTag())
		builder.WriteString(field.Render())
		if field.Field != nil && field.Field.Options.HelpText != "" {
			builder.WriteString(`<div class="helptext">`)
			builder.WriteString(html.EscapeString(field.Field.Options.HelpText))
			builder.WriteString(`</div>`)
		}
		builder.WriteString(`</div>`)
	}
	for _, field := range f.HiddenFields() {
		builder.WriteString(field.Render())
	}
	return builder.String()
}

func (f *Form) RenderErrors() string {
	if f == nil {
		return ""
	}
	f.FullClean()
	if !f.errors.HasErrors() && len(f.nonFieldErrors) == 0 {
		return ""
	}
	var errors ErrorList
	for _, name := range f.fieldOrder {
		label := name
		if field := f.fields[name]; field != nil && field.Options.Label != "" {
			label = field.Options.Label
		}
		for _, err := range f.errors[name] {
			errors = append(errors, ValidationError{Message: label + ": " + err.Message})
		}
	}
	errors = append(errors, ErrorList(f.nonFieldErrors)...)
	return errors.HTML()
}

func (f *Form) boundFieldsMatching(hidden bool) []BoundField {
	if f == nil {
		return nil
	}
	fields := make([]BoundField, 0, len(f.fieldOrder))
	for _, name := range f.fieldOrder {
		boundField := f.BoundField(name)
		if boundField.IsHidden() == hidden {
			fields = append(fields, boundField)
		}
	}
	return fields
}

func (f *Form) htmlName(name string) string {
	if f == nil || f.prefix == "" {
		return name
	}
	return f.prefix + "-" + name
}

func (f *Form) dataValue(name string) any {
	if f == nil {
		return nil
	}
	if value, ok := f.data[f.htmlName(name)]; ok {
		return value
	}
	return f.data[name]
}

func (f *Form) initialValue(name string) any {
	if f == nil {
		return nil
	}
	if value, ok := f.initial[name]; ok {
		return value
	}
	if field := f.fields[name]; field != nil {
		return field.Options.Initial
	}
	return nil
}

// BoundField is one field bound to a form instance.
type BoundField struct {
	Form  *Form
	Name  string
	Field *Field
}

func (bf BoundField) HTMLName() string {
	if bf.Form == nil {
		return bf.Name
	}
	return bf.Form.htmlName(bf.Name)
}

func (bf BoundField) Value() any {
	if bf.Form == nil {
		return nil
	}
	if bf.Form.bound {
		return bf.Form.dataValue(bf.Name)
	}
	return bf.Form.initialValue(bf.Name)
}

func (bf BoundField) Render() string {
	return bf.widget().Render(bf.HTMLName(), bf.Value(), nil)
}

func (bf BoundField) Label() string {
	if bf.Field != nil && bf.Field.Options.Label != "" {
		return bf.Field.Options.Label
	}
	return bf.Name
}

func (bf BoundField) LabelTag() string {
	return `<label for="` + html.EscapeString(bf.HTMLName()) + `">` + html.EscapeString(bf.Label()) + `</label>`
}

func (bf BoundField) Errors() ErrorList {
	if bf.Form == nil {
		return nil
	}
	return bf.Form.FieldErrors(bf.Name)
}

func (bf BoundField) IsHidden() bool {
	return widgetIsHidden(bf.widget())
}

func (bf BoundField) widget() Widget {
	if bf.Field == nil {
		return TextInput()
	}
	if widget, ok := bf.Field.Options.Widget.(Widget); ok && widget != nil {
		return widget
	}
	return defaultWidgetForField(bf.Field)
}

func defaultWidgetForField(field *Field) Widget {
	if field == nil {
		return TextInput()
	}
	switch field.Kind {
	case "boolean":
		return CheckboxInput()
	case "integer", "float", "decimal":
		return NumberInput()
	case "email":
		return EmailInput()
	case "url":
		return URLInput()
	case "date":
		return DateInput()
	case "datetime":
		return DateTimeInput()
	case "time":
		return TimeInput()
	case "file", "image", "multiple_file":
		return FileInput()
	case "choice", "model_choice":
		return Select(field.Choices)
	case "multiple_choice", "model_multiple_choice":
		return SelectMultiple(field.Choices)
	case "split_datetime":
		return SplitDateTimeWidget()
	default:
		return TextInput()
	}
}

func widgetIsHidden(widget Widget) bool {
	switch typed := widget.(type) {
	case inputWidget:
		return typed.inputType == "hidden"
	case multipleHiddenWidget:
		return true
	default:
		return false
	}
}

func cloneFieldMap(fields map[string]*Field) map[string]*Field {
	cloned := make(map[string]*Field, len(fields))
	for name, field := range fields {
		cloned[name] = field
	}
	return cloned
}

func cloneAnyMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneFieldGroups(groups []FieldGroup) []FieldGroup {
	cloned := make([]FieldGroup, len(groups))
	for i, group := range groups {
		cloned[i] = FieldGroup{Name: group.Name, Fields: append([]string(nil), group.Fields...)}
	}
	return cloned
}

func resolveFieldOrder(fields map[string]*Field, requested []string) []string {
	seen := make(map[string]struct{}, len(fields))
	order := make([]string, 0, len(fields))
	for _, name := range requested {
		if _, ok := fields[name]; !ok {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		order = append(order, name)
	}
	remaining := make([]string, 0, len(fields)-len(order))
	for name := range fields {
		if _, ok := seen[name]; ok {
			continue
		}
		remaining = append(remaining, name)
	}
	sort.Strings(remaining)
	return append(order, remaining...)
}

func valuesChanged(left, right any) bool {
	if reflect.DeepEqual(left, right) {
		return false
	}
	if emptyValue(left) && emptyValue(right) {
		return false
	}
	leftStrings, leftErr := toStringSlice(left)
	rightStrings, rightErr := toStringSlice(right)
	if leftErr == nil && rightErr == nil {
		return !reflect.DeepEqual(leftStrings, rightStrings)
	}
	return fmt.Sprint(left) != fmt.Sprint(right)
}
