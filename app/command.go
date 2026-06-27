package app

import (
	"context"
	"fmt"
	"strings"
)

// ManagementCommand is the app-level custom command contract.
type ManagementCommand interface {
	Name() string
	Summary() string
	Run(context.Context, []string) error
}

var reservedCommandNames = map[string]struct{}{
	"help":             {},
	"version":          {},
	"check":            {},
	"runserver":        {},
	"startproject":     {},
	"startapp":         {},
	"makemigrations":   {},
	"migrate":          {},
	"showmigrations":   {},
	"sqlmigrate":       {},
	"squashmigrations": {},
	"createsuperuser":  {},
	"changepassword":   {},
	"collectstatic":    {},
	"shell":            {},
	"dbshell":          {},
	"test":             {},
	"worker":           {},
	"beat":             {},
	"inspect":          {},
	"queues":           {},
	"dumpdata":         {},
	"loaddata":         {},
}

// RegisterManagementCommand registers an app-provided management command.
func (r *Registry) RegisterManagementCommand(command ManagementCommand) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if command == nil {
		return fmt.Errorf("%w: command is nil", ErrInvalidApp)
	}

	name := strings.TrimSpace(command.Name())
	if name == "" {
		return fmt.Errorf("%w: command name is required", ErrInvalidApp)
	}
	if strings.TrimSpace(command.Summary()) == "" {
		return fmt.Errorf("%w: command summary is required", ErrInvalidApp)
	}
	if _, reserved := reservedCommandNames[name]; reserved {
		return fmt.Errorf("%w: %s", ErrReservedCommand, name)
	}
	if _, exists := r.mgmtByName[name]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateCommand, name)
	}

	r.mgmtByName[name] = command
	r.mgmtCommand = append(r.mgmtCommand, command)
	return nil
}

// ManagementCommands returns registered management commands in registration order.
func (r *Registry) ManagementCommands() []ManagementCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]ManagementCommand, len(r.mgmtCommand))
	copy(commands, r.mgmtCommand)
	return commands
}
