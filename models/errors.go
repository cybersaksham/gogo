package models

import "errors"

// ErrMissingStore indicates an operation requires an instance store.
var ErrMissingStore = errors.New("missing instance store")

// ErrInvalidMetadata indicates contradictory or invalid model metadata.
var ErrInvalidMetadata = errors.New("invalid metadata")
