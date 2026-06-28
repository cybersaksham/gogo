package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type TestConfig struct {
	Args   []string
	Stdout io.Writer
	Stderr io.Writer
}

type TestExecutor func(context.Context, TestConfig) error

func NewTestCommand(executor TestExecutor) Command {
	if executor == nil {
		executor = defaultTestExecutor
	}
	return projectTestCommand{executor: executor}
}

type projectTestCommand struct {
	executor TestExecutor
}

func (c projectTestCommand) Name() string {
	return "test"
}

func (c projectTestCommand) Summary() string {
	return "Run project tests"
}

func (c projectTestCommand) Run(ctx context.Context, args []string) error {
	return c.runWithIO(ctx, args, io.Discard, io.Discard)
}

func (c projectTestCommand) runWithIO(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	goTestArgs := append([]string(nil), args...)
	if len(goTestArgs) == 0 {
		goTestArgs = []string{"./..."}
	}
	return c.executor(ctx, TestConfig{
		Args:   goTestArgs,
		Stdout: stdout,
		Stderr: stderr,
	})
}

func defaultTestExecutor(ctx context.Context, config TestConfig) error {
	args := append([]string{"test"}, config.Args...)
	command := exec.CommandContext(ctx, "go", args...)
	if config.Stdout != nil {
		command.Stdout = config.Stdout
	}
	if config.Stderr != nil {
		command.Stderr = config.Stderr
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("%w: go test failed: %v", ErrCommandFailed, err)
	}
	return nil
}
