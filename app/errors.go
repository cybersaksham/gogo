package app

import "errors"

var (
	// ErrInvalidApp indicates invalid app configuration.
	ErrInvalidApp = errors.New("invalid app")

	// ErrDuplicateApp indicates duplicate app names or labels.
	ErrDuplicateApp = errors.New("duplicate app")

	// ErrMissingDependency indicates an app depends on an unregistered app.
	ErrMissingDependency = errors.New("missing app dependency")

	// ErrDependencyCycle indicates app dependencies contain a cycle.
	ErrDependencyCycle = errors.New("app dependency cycle")

	// ErrRegistryReady indicates the registry cannot be modified after readiness.
	ErrRegistryReady = errors.New("app registry ready")
)
