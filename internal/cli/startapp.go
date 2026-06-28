package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	gogotemplates "github.com/cybersaksham/gogo/internal/cli/templates"
)

// NewStartappCommand creates the app generator command.
func NewStartappCommand() Command {
	return startappCommand{}
}

type startappCommand struct{}

func (c startappCommand) Name() string {
	return "startapp"
}

func (c startappCommand) Summary() string {
	return "Create a new Gogo app"
}

func (c startappCommand) Run(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("startapp", flag.ContinueOnError)
	force := flags.Bool("force", false, "allow generation into a non-empty directory")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}

	remaining := flags.Args()
	if len(remaining) < 1 || len(remaining) > 2 {
		return fmt.Errorf("%w: usage startapp [--force] <name> [path]", ErrInvalidArguments)
	}

	appName := remaining[0]
	if !goIdentifierPattern.MatchString(appName) {
		return fmt.Errorf("%w: app name %q must be a valid Go identifier", ErrInvalidArguments, appName)
	}

	target := appName
	if len(remaining) == 2 {
		target = remaining[1]
	}

	if err := ensureProjectTarget(target, *force); err != nil {
		return err
	}

	files, err := gogotemplates.AppFiles(gogotemplates.AppData{AppName: appName, AppLabel: appName})
	if err != nil {
		return fmt.Errorf("%w: render app templates: %v", ErrCommandFailed, err)
	}
	for relativePath, contents := range files {
		path := filepath.Join(target, relativePath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("%w: create directory for %s: %v", ErrCommandFailed, relativePath, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("%w: write %s: %v", ErrCommandFailed, relativePath, err)
		}
	}

	return nil
}
