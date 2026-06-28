package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/cybersaksham/gogo/internal/version"
)

// Root dispatches built-in CLI commands.
type Root struct {
	registry *Registry
}

// NewRoot creates the root CLI with all planned built-in commands.
func NewRoot() *Root {
	registry := NewRegistry()
	root := &Root{registry: registry}

	for _, command := range plannedCommands(root) {
		root.mustRegister(command)
	}

	return root
}

// Execute runs a root command.
func (r *Root) Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		args = []string{"help"}
	}
	if len(args) == 1 {
		switch args[0] {
		case "--help", "-h":
			args = []string{"help"}
		case "--version", "-version":
			args = []string{"version"}
		}
	}

	command, err := r.registry.Get(args[0])
	if err != nil {
		return err
	}

	runArgs := []string(nil)
	if len(args) > 1 {
		runArgs = args[1:]
	}

	runner, ok := command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	})
	if ok {
		return runner.runWithIO(ctx, runArgs, stdout, stderr)
	}

	return command.Run(ctx, runArgs)
}

func (r *Root) mustRegister(command Command) {
	if err := r.registry.Register(command); err != nil {
		panic(err)
	}
}

type helpCommand struct {
	root *Root
}

func (c helpCommand) Name() string {
	return "help"
}

func (c helpCommand) Summary() string {
	return "Show this help message"
}

func (c helpCommand) Run(context.Context, []string) error {
	return nil
}

func (c helpCommand) runWithIO(_ context.Context, _ []string, stdout, _ io.Writer) error {
	_, err := fmt.Fprintln(stdout, "Usage: gogo <command> [args]\n\nCommands:")
	if err != nil {
		return fmt.Errorf("%w: write help: %v", ErrCommandFailed, err)
	}

	for _, command := range c.root.registry.Commands() {
		if _, err := fmt.Fprintf(stdout, "  %-16s %s\n", command.Name(), command.Summary()); err != nil {
			return fmt.Errorf("%w: write help: %v", ErrCommandFailed, err)
		}
	}

	return nil
}

type versionCommand struct{}

func (c versionCommand) Name() string {
	return "version"
}

func (c versionCommand) Summary() string {
	return "Show version metadata"
}

func (c versionCommand) Run(context.Context, []string) error {
	return nil
}

func (c versionCommand) runWithIO(_ context.Context, _ []string, stdout, _ io.Writer) error {
	if _, err := fmt.Fprintln(stdout, version.Info()); err != nil {
		return fmt.Errorf("%w: write version: %v", ErrCommandFailed, err)
	}

	return nil
}

func plannedCommands(root *Root) []Command {
	fixtureStore := NewMemoryFixtureStore()
	return []Command{
		helpCommand{root: root},
		versionCommand{},
		NewCheckCommand(),
		NewRunserverCommand(nil),
		NewStartprojectCommand(),
		NewStartappCommand(),
		NewMakemigrationsCommand(),
		NewMigrateCommand(),
		NewShowmigrationsCommand(),
		NewSQLMigrateCommand(),
		NewSquashmigrationsCommand(),
		NewOptimizeMigrationCommand(),
		NewCreateSuperuserCommand(defaultAuthStore),
		NewChangePasswordCommand(defaultAuthStore),
		NewCollectstaticCommand(nil),
		NewShellCommand(nil),
		NewDBShellCommand(nil),
		NewTestCommand(nil),
		NewWorkerCommand(defaultQueueRuntime),
		NewBeatCommand(defaultQueueRuntime),
		NewInspectCommand(defaultQueueRuntime),
		NewQueuesCommand(defaultQueueRuntime),
		NewDumpdataCommand(fixtureStore),
		NewLoaddataCommand(fixtureStore),
	}
}
