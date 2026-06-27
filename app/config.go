package app

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	importPathPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)
	labelPattern      = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// Registry stores installed app state.
type Registry struct{}

// Config describes a framework app.
type Config interface {
	Name() string
	Label() string
	Path() string
	VerboseName() string
	Dependencies() []string
	Ready(context.Context, *Registry) error
	Shutdown(context.Context) error
}

// BaseConfig is embeddable app configuration for client apps.
type BaseConfig struct {
	AppName         string
	AppLabel        string
	AppPath         string
	AppVerboseName  string
	AppDependencies []string
}

func (c BaseConfig) Name() string {
	return c.AppName
}

func (c BaseConfig) Label() string {
	if c.AppLabel != "" {
		return c.AppLabel
	}

	parts := strings.Split(c.AppName, ".")
	return parts[len(parts)-1]
}

func (c BaseConfig) Path() string {
	return c.AppPath
}

func (c BaseConfig) VerboseName() string {
	if c.AppVerboseName != "" {
		return c.AppVerboseName
	}

	return c.Label()
}

func (c BaseConfig) Dependencies() []string {
	dependencies := make([]string, len(c.AppDependencies))
	copy(dependencies, c.AppDependencies)
	return dependencies
}

func (c BaseConfig) Ready(context.Context, *Registry) error {
	return nil
}

func (c BaseConfig) Shutdown(context.Context) error {
	return nil
}

// ValidateConfig validates one app config.
func ValidateConfig(config Config) error {
	if config == nil {
		return fmt.Errorf("%w: config is nil", ErrInvalidApp)
	}

	if !importPathPattern.MatchString(config.Name()) {
		return fmt.Errorf("%w: app name %q must be an import-like dotted path", ErrInvalidApp, config.Name())
	}

	if !labelPattern.MatchString(config.Label()) {
		return fmt.Errorf("%w: app label %q must be a valid identifier", ErrInvalidApp, config.Label())
	}

	if strings.TrimSpace(config.Path()) == "" {
		return fmt.Errorf("%w: app path is required for %s", ErrInvalidApp, config.Name())
	}

	for _, dependency := range config.Dependencies() {
		if !importPathPattern.MatchString(dependency) {
			return fmt.Errorf("%w: dependency %q must be an import-like dotted path", ErrInvalidApp, dependency)
		}
	}

	return nil
}

// ValidateConfigs validates app configs together for cross-app constraints.
func ValidateConfigs(configs ...Config) error {
	labels := make(map[string]string, len(configs))
	for _, config := range configs {
		if err := ValidateConfig(config); err != nil {
			return err
		}

		if existingName, exists := labels[config.Label()]; exists {
			return fmt.Errorf("%w: label %q is used by %s and %s", ErrDuplicateApp, config.Label(), existingName, config.Name())
		}
		labels[config.Label()] = config.Name()
	}

	return nil
}
