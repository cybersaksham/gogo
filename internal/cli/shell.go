package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/cybersaksham/gogo/app"
	"github.com/cybersaksham/gogo/conf"
)

// ShellConfig contains resolved shell execution context.
type ShellConfig struct {
	Command  string
	Settings conf.Settings
	Registry *app.Registry
}

// ShellExecutor executes shell commands with resolved project context.
type ShellExecutor func(context.Context, ShellConfig) error

// NewShellCommand creates the shell command.
func NewShellCommand(executor ShellExecutor) Command {
	if executor == nil {
		executor = unavailableShellExecutor
	}
	return shellCommand{executor: executor}
}

type shellCommand struct {
	executor ShellExecutor
}

func (c shellCommand) Name() string {
	return "shell"
}

func (c shellCommand) Summary() string {
	return "Open a project shell"
}

func (c shellCommand) Run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("shell", flag.ContinueOnError)
	command := flags.String("command", "", "non-interactive command to execute")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}

	settings, err := conf.LoadFromEnv()
	if err != nil {
		return err
	}
	if err := settings.Validate(); err != nil {
		return err
	}

	return c.executor(ctx, ShellConfig{
		Command:  *command,
		Settings: settings,
		Registry: app.NewRegistry(),
	})
}

func unavailableShellExecutor(context.Context, ShellConfig) error {
	return fmt.Errorf("%w: interactive shell execution is planned for phase 02 app registry integration", ErrCommandUnavailable)
}
