package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cybersaksham/gogo/app"
	"github.com/cybersaksham/gogo/conf"
)

// ShellConfig contains resolved shell execution context.
type ShellConfig struct {
	Command  string
	Settings conf.Settings
	Registry *app.Registry
	Stdout   io.Writer
	Stderr   io.Writer
}

// ShellExecutor executes shell commands with resolved project context.
type ShellExecutor func(context.Context, ShellConfig) error

// NewShellCommand creates the shell command.
func NewShellCommand(executor ShellExecutor) Command {
	if executor == nil {
		executor = defaultShellExecutor
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
	return c.runWithIO(ctx, args, io.Discard, io.Discard)
}

func (c shellCommand) runWithIO(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("shell", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
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
		Stdout:   stdout,
		Stderr:   stderr,
	})
}

func defaultShellExecutor(ctx context.Context, config ShellConfig) error {
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		shell = "sh"
	}

	args := []string(nil)
	if config.Command != "" {
		args = []string{"-c", config.Command}
	}

	command := exec.CommandContext(ctx, shell, args...)
	if config.Command == "" {
		command.Stdin = os.Stdin
	}
	if config.Stdout != nil {
		command.Stdout = config.Stdout
	}
	if config.Stderr != nil {
		command.Stderr = config.Stderr
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("%w: shell command failed: %v", ErrCommandFailed, err)
	}
	return nil
}
