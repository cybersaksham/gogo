package http

import "errors"

// ErrInvalidPattern indicates an invalid route pattern.
var ErrInvalidPattern = errors.New("invalid pattern")

// ErrRouteConflict indicates a duplicate route name or method conflict.
var ErrRouteConflict = errors.New("route conflict")
