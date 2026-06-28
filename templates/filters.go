package templates

import (
	"fmt"
	"html"
	"html/template"
	"reflect"
	"strings"
	"time"
)

func TemplateFilters() template.FuncMap {
	return template.FuncMap{
		"date":        formatDate,
		"default":     defaultValue,
		"length":      lengthOf,
		"join":        joinValues,
		"pluralize":   pluralize,
		"linebreaks":  linebreaks,
		"safe_escape": safeEscape,
		"escape":      safeEscape,
	}
}

func formatDate(value any, layout string) string {
	if layout == "" {
		layout = "2006-01-02"
	}
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return ""
		}
		return typed.Format(layout)
	case *time.Time:
		if typed == nil || typed.IsZero() {
			return ""
		}
		return typed.Format(layout)
	default:
		return fmt.Sprint(value)
	}
}

func defaultValue(value any, fallback any) any {
	if helperEmpty(value) {
		return fallback
	}
	return value
}

func lengthOf(value any) int {
	if value == nil {
		return 0
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return reflected.Len()
	default:
		return 0
	}
}

func joinValues(value any, separator string) string {
	if value == nil {
		return ""
	}
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Array && reflected.Kind() != reflect.Slice {
		return fmt.Sprint(value)
	}
	parts := make([]string, reflected.Len())
	for i := 0; i < reflected.Len(); i++ {
		parts[i] = fmt.Sprint(reflected.Index(i).Interface())
	}
	return strings.Join(parts, separator)
}

func pluralize(count any, singular string, plural ...string) string {
	if numericValue(count) == 1 {
		return singular
	}
	if len(plural) > 0 {
		return plural[0]
	}
	return singular + "s"
}

func linebreaks(value any) SafeString {
	escaped := html.EscapeString(fmt.Sprint(value))
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	return SafeString("<p>" + strings.ReplaceAll(escaped, "\n", "<br>") + "</p>")
}

func safeEscape(value any) SafeString {
	return SafeString(html.EscapeString(fmt.Sprint(value)))
}

func helperEmpty(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Bool:
		return !reflected.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflected.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflected.Float() == 0
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return reflected.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return reflected.IsNil()
	default:
		return false
	}
}

func numericValue(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case uint:
		return int64(typed)
	case uint8:
		return int64(typed)
	case uint16:
		return int64(typed)
	case uint32:
		return int64(typed)
	case uint64:
		return int64(typed)
	default:
		return int64(lengthOf(value))
	}
}
