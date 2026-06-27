package app

import "errors"

var (
	// ErrInvalidApp indicates invalid app configuration.
	ErrInvalidApp = errors.New("invalid app")

	// ErrDuplicateApp indicates duplicate app names or labels.
	ErrDuplicateApp = errors.New("duplicate app")
)
