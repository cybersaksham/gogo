package cli

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRegistryReturnsCommandsInRegistrationOrder(t *testing.T) {
	registry := NewRegistry()

	mustRegister(t, registry, testCommand{name: "version", summary: "Show version"})
	mustRegister(t, registry, testCommand{name: "check", summary: "Run checks"})
	mustRegister(t, registry, testCommand{name: "runserver", summary: "Run server"})

	got := commandNames(registry.Commands())
	want := []string{"version", "check", "runserver"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Commands() names = %#v, want %#v", got, want)
	}
}

func TestRegistryRejectsDuplicateCommands(t *testing.T) {
	registry := NewRegistry()

	mustRegister(t, registry, testCommand{name: "check", summary: "Run checks"})

	err := registry.Register(testCommand{name: "check", summary: "Run checks again"})
	if !errors.Is(err, ErrDuplicateCommand) {
		t.Fatalf("Register duplicate error = %v, want ErrDuplicateCommand", err)
	}
}

func TestRegistryLooksUpRegisteredCommand(t *testing.T) {
	registry := NewRegistry()
	command := testCommand{name: "check", summary: "Run checks"}

	mustRegister(t, registry, command)

	got, err := registry.Get("check")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name() != "check" {
		t.Fatalf("Get() command = %q, want %q", got.Name(), "check")
	}
}

func TestRegistryReturnsUnknownCommandError(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("missing")
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("Get() error = %v, want ErrUnknownCommand", err)
	}
}

func TestRegistryRejectsInvalidCommands(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(testCommand{name: "", summary: "No name"})
	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("Register empty name error = %v, want ErrInvalidArguments", err)
	}

	err = registry.Register(testCommand{name: "check", summary: ""})
	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("Register empty summary error = %v, want ErrInvalidArguments", err)
	}
}

func mustRegister(t *testing.T, registry *Registry, command Command) {
	t.Helper()

	if err := registry.Register(command); err != nil {
		t.Fatalf("Register(%q) error = %v", command.Name(), err)
	}
}

func commandNames(commands []Command) []string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.Name())
	}
	return names
}

type testCommand struct {
	name    string
	summary string
}

func (c testCommand) Name() string {
	return c.name
}

func (c testCommand) Summary() string {
	return c.summary
}

func (c testCommand) Run(context.Context, []string) error {
	return nil
}
