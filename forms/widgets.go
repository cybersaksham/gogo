package forms

import (
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Attrs contains HTML attributes applied to a widget render call.
type Attrs map[string]string

// Widget renders one form value into HTML.
type Widget interface {
	Render(name string, value any, attrs Attrs) string
}

type inputWidget struct {
	inputType      string
	includeValue   bool
	valueFormatter func(any) string
}

func TextInput() Widget {
	return inputWidget{inputType: "text", includeValue: true, valueFormatter: formatWidgetValue}
}

func NumberInput() Widget {
	return inputWidget{inputType: "number", includeValue: true, valueFormatter: formatWidgetValue}
}

func EmailInput() Widget {
	return inputWidget{inputType: "email", includeValue: true, valueFormatter: formatWidgetValue}
}

func URLInput() Widget {
	return inputWidget{inputType: "url", includeValue: true, valueFormatter: formatWidgetValue}
}

func PasswordInput() Widget {
	return inputWidget{inputType: "password", includeValue: true, valueFormatter: func(any) string { return "" }}
}

func HiddenInput() Widget {
	return inputWidget{inputType: "hidden", includeValue: true, valueFormatter: formatWidgetValue}
}

func DateInput() Widget {
	return inputWidget{inputType: "date", includeValue: true, valueFormatter: formatDateValue}
}

func DateTimeInput() Widget {
	return inputWidget{inputType: "datetime-local", includeValue: true, valueFormatter: formatDateTimeValue}
}

func TimeInput() Widget {
	return inputWidget{inputType: "time", includeValue: true, valueFormatter: formatTimeValue}
}

func FileInput() Widget {
	return inputWidget{inputType: "file"}
}

func (w inputWidget) Render(name string, value any, attrs Attrs) string {
	renderAttrs := cloneAttrs(attrs)
	renderAttrs["name"] = name
	renderAttrs["type"] = w.inputType
	if w.includeValue {
		formatter := w.valueFormatter
		if formatter == nil {
			formatter = formatWidgetValue
		}
		renderAttrs["value"] = formatter(value)
	}
	return "<input" + renderHTMLAttrs(renderAttrs, nil) + ">"
}

type multipleHiddenWidget struct{}

func MultipleHiddenInput() Widget {
	return multipleHiddenWidget{}
}

func (multipleHiddenWidget) Render(name string, value any, attrs Attrs) string {
	values := widgetValueSlice(value)
	rendered := make([]string, 0, len(values))
	for _, item := range values {
		rendered = append(rendered, HiddenInput().Render(name, item, attrs))
	}
	return strings.Join(rendered, "\n")
}

type textareaWidget struct{}

func Textarea() Widget {
	return textareaWidget{}
}

func (textareaWidget) Render(name string, value any, attrs Attrs) string {
	renderAttrs := cloneAttrs(attrs)
	renderAttrs["name"] = name
	return "<textarea" + renderHTMLAttrs(renderAttrs, nil) + ">" + html.EscapeString(formatWidgetValue(value)) + "</textarea>"
}

type checkboxWidget struct{}

func CheckboxInput() Widget {
	return checkboxWidget{}
}

func (checkboxWidget) Render(name string, value any, attrs Attrs) string {
	renderAttrs := cloneAttrs(attrs)
	renderAttrs["name"] = name
	renderAttrs["type"] = "checkbox"
	renderAttrs["value"] = "true"
	boolAttrs := map[string]bool{"checked": widgetBool(value)}
	return "<input" + renderHTMLAttrs(renderAttrs, boolAttrs) + ">"
}

type selectWidget struct {
	choices  []Choice
	multiple bool
}

func Select(choices []Choice) Widget {
	return selectWidget{choices: cloneChoices(choices)}
}

func SelectMultiple(choices []Choice) Widget {
	return selectWidget{choices: cloneChoices(choices), multiple: true}
}

func (w selectWidget) Render(name string, value any, attrs Attrs) string {
	renderAttrs := cloneAttrs(attrs)
	renderAttrs["name"] = name
	boolAttrs := map[string]bool{"multiple": w.multiple}
	selected := widgetSelectedSet(value, w.multiple)
	var builder strings.Builder
	builder.WriteString("<select")
	builder.WriteString(renderHTMLAttrs(renderAttrs, boolAttrs))
	builder.WriteString(">")
	for _, choice := range w.choices {
		choiceValue := formatWidgetValue(choice.Value)
		optionAttrs := Attrs{"value": choiceValue}
		optionBools := map[string]bool{"selected": selected[choiceValue]}
		builder.WriteString("<option")
		builder.WriteString(renderHTMLAttrs(optionAttrs, optionBools))
		builder.WriteString(">")
		builder.WriteString(html.EscapeString(choiceLabel(choice)))
		builder.WriteString("</option>")
	}
	builder.WriteString("</select>")
	return builder.String()
}

type choiceListWidget struct {
	inputType string
	choices   []Choice
	multiple  bool
}

func RadioSelect(choices []Choice) Widget {
	return choiceListWidget{inputType: "radio", choices: cloneChoices(choices)}
}

func CheckboxSelectMultiple(choices []Choice) Widget {
	return choiceListWidget{inputType: "checkbox", choices: cloneChoices(choices), multiple: true}
}

func (w choiceListWidget) Render(name string, value any, attrs Attrs) string {
	selected := widgetSelectedSet(value, w.multiple)
	rendered := make([]string, 0, len(w.choices))
	for _, choice := range w.choices {
		choiceValue := formatWidgetValue(choice.Value)
		inputAttrs := cloneAttrs(attrs)
		inputAttrs["name"] = name
		inputAttrs["type"] = w.inputType
		inputAttrs["value"] = choiceValue
		boolAttrs := map[string]bool{"checked": selected[choiceValue]}
		rendered = append(rendered, "<label><input"+renderHTMLAttrs(inputAttrs, boolAttrs)+"> "+html.EscapeString(choiceLabel(choice))+"</label>")
	}
	return strings.Join(rendered, "\n")
}

type clearableFileWidget struct{}

func ClearableFileInput() Widget {
	return clearableFileWidget{}
}

func (clearableFileWidget) Render(name string, value any, attrs Attrs) string {
	current := currentFileName(value)
	input := FileInput().Render(name, nil, attrs)
	if current == "" {
		return input
	}
	clearAttrs := Attrs{
		"name":  name + "-clear",
		"type":  "checkbox",
		"value": "true",
	}
	return `<span class="current-file">` + html.EscapeString(current) + `</span><label><input` + renderHTMLAttrs(clearAttrs, nil) + `> Clear</label>` + input
}

type splitDateTimeWidget struct{}

func SplitDateTimeWidget() Widget {
	return splitDateTimeWidget{}
}

func (splitDateTimeWidget) Render(name string, value any, attrs Attrs) string {
	dateValue, timeValue := splitDateTimeWidgetValues(value)
	return DateInput().Render(name+"_0", dateValue, attrs) + TimeInput().Render(name+"_1", timeValue, attrs)
}

var safeAttrNameRegexp = regexp.MustCompile(`^[A-Za-z_:][A-Za-z0-9_:.-]*$`)

func renderHTMLAttrs(attrs Attrs, boolAttrs map[string]bool) string {
	keys := make([]string, 0, len(attrs)+len(boolAttrs))
	seen := make(map[string]struct{}, len(attrs)+len(boolAttrs))
	for key := range attrs {
		if !safeAttrNameRegexp.MatchString(key) {
			continue
		}
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key, enabled := range boolAttrs {
		if !enabled || !safeAttrNameRegexp.MatchString(key) {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteByte(' ')
		builder.WriteString(html.EscapeString(key))
		if boolAttrs[key] {
			continue
		}
		builder.WriteString(`="`)
		builder.WriteString(html.EscapeString(attrs[key]))
		builder.WriteByte('"')
	}
	return builder.String()
}

func cloneAttrs(attrs Attrs) Attrs {
	cloned := make(Attrs, len(attrs)+3)
	for key, value := range attrs {
		cloned[key] = value
	}
	return cloned
}

func formatWidgetValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func formatDateValue(value any) string {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return ""
		}
		return typed.Format("2006-01-02")
	default:
		return formatWidgetValue(value)
	}
}

func formatDateTimeValue(value any) string {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return ""
		}
		return typed.Format("2006-01-02T15:04:05")
	default:
		return formatWidgetValue(value)
	}
}

func formatTimeValue(value any) string {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return ""
		}
		return typed.Format("15:04:05")
	default:
		return formatWidgetValue(value)
	}
}

func widgetValueSlice(value any) []string {
	if emptyValue(value) {
		return nil
	}
	values, err := toStringSlice(value)
	if err == nil {
		return values
	}
	return []string{formatWidgetValue(value)}
}

func widgetSelectedSet(value any, multiple bool) map[string]bool {
	selected := make(map[string]bool)
	if multiple {
		for _, item := range widgetValueSlice(value) {
			selected[item] = true
		}
		return selected
	}
	if !emptyValue(value) {
		selected[formatWidgetValue(value)] = true
	}
	return selected
}

func widgetBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "on", "yes":
			return true
		default:
			return false
		}
	default:
		return !emptyValue(value)
	}
}

func choiceLabel(choice Choice) string {
	if choice.Label != "" {
		return choice.Label
	}
	return formatWidgetValue(choice.Value)
}

func currentFileName(value any) string {
	switch typed := value.(type) {
	case UploadedFile:
		return typed.Name
	case *UploadedFile:
		if typed == nil {
			return ""
		}
		return typed.Name
	default:
		if emptyValue(value) {
			return ""
		}
		return formatWidgetValue(value)
	}
}

func splitDateTimeWidgetValues(value any) (any, any) {
	switch typed := value.(type) {
	case time.Time:
		return typed, typed
	default:
		values, err := toAnySlice(value)
		if err == nil && len(values) == 2 {
			return values[0], values[1]
		}
		return value, nil
	}
}
