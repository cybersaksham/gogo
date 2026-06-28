package forms

import (
	"errors"
	"html"
	"strings"
)

// ValidationError is a user-facing form validation error.
type ValidationError struct {
	Code    string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// ErrorList is an ordered list of validation errors.
type ErrorList []ValidationError

func (l ErrorList) Messages() []string {
	messages := make([]string, len(l))
	for i, err := range l {
		messages[i] = err.Message
	}
	return messages
}

func (l ErrorList) HTML() string {
	return l.html("errorlist")
}

func (l ErrorList) html(className string) string {
	if len(l) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(`<ul class="`)
	builder.WriteString(html.EscapeString(className))
	builder.WriteString(`">`)
	for _, err := range l {
		builder.WriteString("<li>")
		builder.WriteString(html.EscapeString(err.Message))
		builder.WriteString("</li>")
	}
	builder.WriteString("</ul>")
	return builder.String()
}

// NonFieldErrorList renders form-level errors with a distinct CSS class.
type NonFieldErrorList ErrorList

func (l NonFieldErrorList) Messages() []string {
	return ErrorList(l).Messages()
}

func (l NonFieldErrorList) HTML() string {
	return ErrorList(l).html("errorlist nonfield")
}

// ErrorDict stores field-specific validation errors.
type ErrorDict map[string]ErrorList

func (d ErrorDict) Add(field string, err error) {
	if err == nil {
		return
	}
	d[field] = append(d[field], normalizeValidationError(err))
}

func (d ErrorDict) Get(field string) ErrorList {
	return append(ErrorList(nil), d[field]...)
}

func (d ErrorDict) HasErrors() bool {
	for _, errors := range d {
		if len(errors) > 0 {
			return true
		}
	}
	return false
}

func normalizeValidationError(err error) ValidationError {
	var validationError ValidationError
	if errors.As(err, &validationError) {
		return validationError
	}
	message := err.Error()
	if errors.Is(err, ErrValidation) {
		message = strings.TrimPrefix(message, ErrValidation.Error()+": ")
	}
	return ValidationError{Message: message}
}
