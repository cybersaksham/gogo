package cli

import (
	"context"
	"fmt"
	"strings"
)

// Command is the common contract for built-in and app-provided commands.
type Command interface {
	Name() string
	Summary() string
	Run(context.Context, []string) error
}

// Registry stores commands in deterministic registration order.
type Registry struct {
	commands []Command
	byName   map[string]Command
}

// NewRegistry creates an empty command registry.
func NewRegistry() *Registry {
	return &Registry{
		byName: make(map[string]Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(command Command) error {
	if command == nil {
		return fmt.Errorf("%w: command is nil", ErrInvalidArguments)
	}

	name := strings.TrimSpace(command.Name())
	if name == "" {
		return fmt.Errorf("%w: command name is required", ErrInvalidArguments)
	}

	if strings.TrimSpace(command.Summary()) == "" {
		return fmt.Errorf("%w: command summary is required", ErrInvalidArguments)
	}

	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateCommand, name)
	}

	r.byName[name] = command
	r.commands = append(r.commands, command)
	return nil
}

// Get returns a registered command by name.
func (r *Registry) Get(name string) (Command, error) {
	command, exists := r.byName[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCommand, name)
	}

	return command, nil
}

// Commands returns registered commands in registration order.
func (r *Registry) Commands() []Command {
	commands := make([]Command, len(r.commands))
	copy(commands, r.commands)
	return commands
}
