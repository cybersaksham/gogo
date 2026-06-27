package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/cybersaksham/gogo/conf"
)

// NewCheckCommand creates the built-in system check command.
func NewCheckCommand() Command {
	return checkCommand{}
}

type checkCommand struct{}

func (c checkCommand) Name() string {
	return "check"
}

func (c checkCommand) Summary() string {
	return "Run system checks"
}

func (c checkCommand) Run(context.Context, []string) error {
	settings, err := conf.LoadFromEnv()
	if err != nil {
		return err
	}
	return settings.Validate()
}

func (c checkCommand) runWithIO(_ context.Context, _ []string, stdout, _ io.Writer) error {
	settings, err := conf.LoadFromEnv()
	if err != nil {
		return err
	}

	if err := settings.Validate(); err != nil {
		fmt.Fprintf(stdout, "ERROR config %v\n", err)
		return err
	}

	fmt.Fprintln(stdout, "OK config settings valid")
	fmt.Fprintln(stdout, "WARN apps app registry checks unavailable until phase 02-app-project-lifecycle")

	return nil
}
