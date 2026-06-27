package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	for relativePath, contents := range appFiles(appName) {
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

func appFiles(appName string) map[string]string {
	return map[string]string{
		"app.go":                             appPackageFile(appName, "App configuration."),
		"models.go":                          appPackageFile(appName, "Models."),
		"admin.go":                           appPackageFile(appName, "Admin registrations."),
		"urls.go":                            appPackageFile(appName, "Routes."),
		"api.go":                             appPackageFile(appName, "API routes."),
		"serializers.go":                     appPackageFile(appName, "Serializers."),
		"forms.go":                           appPackageFile(appName, "Forms."),
		"services.go":                        appPackageFile(appName, "Application services."),
		"tasks.go":                           appPackageFile(appName, "Queue tasks."),
		"permissions.go":                     appPackageFile(appName, "Permissions."),
		filepath.Join("migrations", ".keep"): "",
		filepath.Join("templates", appName, ".keep"): "",
		filepath.Join("static", appName, ".keep"):    "",
		filepath.Join("tests", ".keep"):              "",
	}
}

func appPackageFile(packageName string, comment string) string {
	return fmt.Sprintf("// %s\npackage %s\n", comment, packageName)
}
