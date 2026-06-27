package fields

import "errors"

// ErrValidation indicates a field value failed validation.
var ErrValidation = errors.New("field validation")

// ErrInvalidField indicates invalid field configuration.
var ErrInvalidField = errors.New("invalid field")
