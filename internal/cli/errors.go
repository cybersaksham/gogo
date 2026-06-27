package cli

import "errors"

var (
	// ErrUnknownCommand indicates that a command name was not registered.
	ErrUnknownCommand = errors.New("unknown command")

	// ErrDuplicateCommand indicates that a command name was registered twice.
	ErrDuplicateCommand = errors.New("duplicate command")

	// ErrInvalidArguments indicates invalid CLI command arguments or metadata.
	ErrInvalidArguments = errors.New("invalid arguments")

	// ErrCommandFailed indicates a command failed while running.
	ErrCommandFailed = errors.New("command failed")
)
