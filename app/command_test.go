package app

import (
	"context"
	"errors"
	"testing"
)

func TestRegistryRegistersAndExecutesManagementCommand(t *testing.T) {
	registry := NewRegistry()
	command := &testManagementCommand{name: "blog.reindex", summary: "Reindex blog"}

	if err := registry.RegisterManagementCommand(command); err != nil {
		t.Fatalf("RegisterManagementCommand() error = %v", err)
	}

	commands := registry.ManagementCommands()
	if len(commands) != 1 {
		t.Fatalf("ManagementCommands() length = %d, want 1", len(commands))
	}

	if err := commands[0].Run(context.Background(), []string{"--all"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !command.called {
		t.Fatalf("command was not executed")
	}
}

func TestRegistryRejectsDuplicateManagementCommand(t *testing.T) {
	registry := NewRegistry()
	if err := registry.RegisterManagementCommand(&testManagementCommand{name: "blog.reindex", summary: "Reindex"}); err != nil {
		t.Fatalf("RegisterManagementCommand() error = %v", err)
	}

	err := registry.RegisterManagementCommand(&testManagementCommand{name: "blog.reindex", summary: "Reindex again"})
	if !errors.Is(err, ErrDuplicateCommand) {
		t.Fatalf("RegisterManagementCommand() error = %v, want ErrDuplicateCommand", err)
	}
}

func TestRegistryRejectsBuiltInCommandNameCollision(t *testing.T) {
	registry := NewRegistry()

	err := registry.RegisterManagementCommand(&testManagementCommand{name: "check", summary: "Custom check"})
	if !errors.Is(err, ErrReservedCommand) {
		t.Fatalf("RegisterManagementCommand() error = %v, want ErrReservedCommand", err)
	}
}

func TestRegistryAllowsNamespacedCommandThatContainsBuiltInName(t *testing.T) {
	registry := NewRegistry()

	err := registry.RegisterManagementCommand(&testManagementCommand{name: "blog.check", summary: "Blog check"})
	if err != nil {
		t.Fatalf("RegisterManagementCommand() error = %v", err)
	}
}

func TestManagementCommandsReturnsCopy(t *testing.T) {
	registry := NewRegistry()
	if err := registry.RegisterManagementCommand(&testManagementCommand{name: "blog.reindex", summary: "Reindex"}); err != nil {
		t.Fatalf("RegisterManagementCommand() error = %v", err)
	}

	commands := registry.ManagementCommands()
	commands[0] = &testManagementCommand{name: "changed", summary: "Changed"}

	if got := registry.ManagementCommands()[0].Name(); got != "blog.reindex" {
		t.Fatalf("ManagementCommands() leaked internal slice, got %q", got)
	}
}

type testManagementCommand struct {
	name    string
	summary string
	called  bool
}

func (c *testManagementCommand) Name() string {
	return c.name
}

func (c *testManagementCommand) Summary() string {
	return c.summary
}

func (c *testManagementCommand) Run(context.Context, []string) error {
	c.called = true
	return nil
}
