package admin

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

// WidgetChoice describes one select option.
type WidgetChoice struct {
	Value string
	Label string
}

// WidgetConfig configures an admin widget render.
type WidgetConfig struct {
	Name        string
	Value       any
	Attrs       map[string]string
	Choices     []WidgetChoice
	RelationURL string
}

// TextInput renders a text input.
func TextInput(config WidgetConfig) string {
	return input("text", config, nil)
}

// Textarea renders a textarea.
func Textarea(config WidgetConfig) string {
	return fmt.Sprintf(`<textarea name="%s"%s>%s</textarea>`, esc(config.Name), renderAttrs(config.Attrs), esc(fmt.Sprint(config.Value)))
}

// NumberInput renders a number input.
func NumberInput(config WidgetConfig) string {
	return input("number", config, nil)
}

// Checkbox renders a checkbox input.
func Checkbox(config WidgetConfig) string {
	extra := map[string]string{}
	if checked, _ := config.Value.(bool); checked {
		extra["checked"] = "checked"
	}
	return input("checkbox", config, extra)
}

// Select renders a select input.
func Select(config WidgetConfig) string {
	return selectWidget(config, false, selectedValues(config.Value))
}

// SelectMultiple renders a multiple select input.
func SelectMultiple(config WidgetConfig) string {
	return selectWidget(config, true, selectedValues(config.Value))
}

// DateInput renders a date input.
func DateInput(config WidgetConfig) string {
	return input("date", config, nil)
}

// TimeInput renders a time input.
func TimeInput(config WidgetConfig) string {
	return input("time", config, nil)
}

// DateTimeInput renders a datetime-local input.
func DateTimeInput(config WidgetConfig) string {
	return input("datetime-local", config, nil)
}

// FileInput renders a file input.
func FileInput(config WidgetConfig) string {
	config.Value = ""
	return input("file", config, nil)
}

// ClearableFileInput renders a file input with a clear checkbox.
func ClearableFileInput(config WidgetConfig) string {
	current := ""
	if fmt.Sprint(config.Value) != "" {
		current = `<span class="current-file">` + esc(fmt.Sprint(config.Value)) + `</span>`
	}
	return current + FileInput(config) + fmt.Sprintf(`<label><input type="checkbox" name="%s-clear" value="1"> Clear</label>`, esc(config.Name))
}

// RawIDRelationWidget renders a raw ID relation input.
func RawIDRelationWidget(config WidgetConfig) string {
	return input("text", config, map[string]string{"data-lookup-url": config.RelationURL})
}

// AutocompleteWidget renders an autocomplete relation input.
func AutocompleteWidget(config WidgetConfig) string {
	return input("text", config, map[string]string{"data-autocomplete-url": config.RelationURL})
}

// FilteredSelectMultiple renders a Django-style filtered multiple select.
func FilteredSelectMultiple(config WidgetConfig) string {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["class"] = strings.TrimSpace(attrs["class"] + " filtered-select-multiple")
	config.Attrs = attrs
	return SelectMultiple(config)
}

// ReadonlyDisplay renders an escaped readonly value.
func ReadonlyDisplay(config WidgetConfig) string {
	return fmt.Sprintf(`<span class="readonly" data-field="%s">%s</span>`, esc(config.Name), esc(fmt.Sprint(config.Value)))
}

func input(inputType string, config WidgetConfig, extra map[string]string) string {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	for key, value := range extra {
		attrs[key] = value
	}
	return fmt.Sprintf(`<input type="%s" name="%s" value="%s"%s>`, esc(inputType), esc(config.Name), esc(fmt.Sprint(config.Value)), renderAttrs(attrs))
}

func selectWidget(config WidgetConfig, multiple bool, selected map[string]struct{}) string {
	attrs := cloneStringMap(config.Attrs)
	if attrs == nil {
		attrs = map[string]string{}
	}
	if multiple {
		attrs["multiple"] = "multiple"
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`<select name="%s"%s>`, esc(config.Name), renderAttrs(attrs)))
	for _, choice := range config.Choices {
		selectedAttr := ""
		if _, ok := selected[choice.Value]; ok {
			selectedAttr = " selected"
		}
		builder.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, esc(choice.Value), selectedAttr, esc(choice.Label)))
	}
	builder.WriteString(`</select>`)
	return builder.String()
}

func selectedValues(value any) map[string]struct{} {
	result := map[string]struct{}{}
	switch typed := value.(type) {
	case []string:
		for _, item := range typed {
			result[item] = struct{}{}
		}
	default:
		if value != nil {
			result[fmt.Sprint(value)] = struct{}{}
		}
	}
	return result
}

func renderAttrs(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		if attrs[key] == key {
			builder.WriteString(" ")
			builder.WriteString(esc(key))
			continue
		}
		builder.WriteString(fmt.Sprintf(` %s="%s"`, esc(key), esc(attrs[key])))
	}
	return builder.String()
}

func esc(value string) string {
	return html.EscapeString(value)
}
