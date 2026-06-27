package validation

import "fmt"

// Error is one structured validation error.
type Error struct {
	Code    string
	Message string
}

// Errors stores validation errors keyed by field name.
type Errors struct {
	Fields map[string][]Error
}

func (e *Errors) Error() string {
	return fmt.Sprintf("validation errors: %v", e.Fields)
}

// Add appends one field error.
func (e *Errors) Add(field, code, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string][]Error)
	}
	e.Fields[field] = append(e.Fields[field], Error{Code: code, Message: message})
}

// Has reports whether a field has any errors.
func (e *Errors) Has(field string) bool {
	return len(e.Fields[field]) > 0
}

func (e *Errors) empty() bool {
	return len(e.Fields) == 0
}
