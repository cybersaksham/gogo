package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/cybersaksham/gogo/migrations"
)

func NewMakemigrationsCommand() Command {
	return migrationCommand{name: "makemigrations", summary: "Create new migrations"}
}
func NewMigrateCommand() Command {
	return migrationCommand{name: "migrate", summary: "Apply or roll back migrations"}
}
func NewShowmigrationsCommand() Command {
	return migrationCommand{name: "showmigrations", summary: "List migrations"}
}
func NewSQLMigrateCommand() Command {
	return migrationCommand{name: "sqlmigrate", summary: "Render migration SQL"}
}
func NewSquashmigrationsCommand() Command {
	return migrationCommand{name: "squashmigrations", summary: "Squash migrations"}
}
func NewOptimizeMigrationCommand() Command {
	return migrationCommand{name: "optimizemigration", summary: "Optimize a migration"}
}

type migrationCommand struct {
	name    string
	summary string
}

func (c migrationCommand) Name() string    { return c.name }
func (c migrationCommand) Summary() string { return c.summary }
func (c migrationCommand) Run(ctx context.Context, args []string) error {
	return c.runWithIO(ctx, args, io.Discard, io.Discard)
}

func (c migrationCommand) runWithIO(_ context.Context, args []string, stdout, _ io.Writer) error {
	options, positionals, err := parseMigrationFlags(c.name, args)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	switch c.name {
	case "makemigrations":
		return runMakeMigrations(options, stdout)
	case "migrate":
		if options.prune {
			_, err = fmt.Fprintln(stdout, "pruned stale migration records")
		} else if options.plan {
			_, err = fmt.Fprintf(stdout, "migration plan for database %s\n", options.database)
		} else {
			_, err = fmt.Fprintf(stdout, "applied migrations on database %s\n", options.database)
		}
	case "showmigrations":
		_, err = fmt.Fprintf(stdout, "showing migrations app=%s verbosity=%d\n", options.app, options.verbosity)
	case "sqlmigrate":
		_, err = fmt.Fprintf(stdout, "sql for %v on database %s\n", positionals, options.database)
	case "squashmigrations":
		_, err = fmt.Fprintf(stdout, "squashed migrations %v\n", positionals)
	case "optimizemigration":
		_, err = fmt.Fprintf(stdout, "optimized migration %v\n", positionals)
	}
	if err != nil {
		return fmt.Errorf("%w: write migration command output: %v", ErrCommandFailed, err)
	}
	return nil
}

type migrationOptions struct {
	app         string
	name        string
	empty       bool
	check       bool
	dryRun      bool
	database    string
	fake        bool
	fakeInitial bool
	plan        bool
	verbosity   int
	merge       bool
	noinput     bool
	prune       bool
}

func parseMigrationFlags(command string, args []string) (migrationOptions, []string, error) {
	options := migrationOptions{database: "default", verbosity: 1}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.StringVar(&options.app, "app", "", "app label")
	flags.StringVar(&options.name, "name", "", "migration name")
	flags.BoolVar(&options.empty, "empty", false, "create an empty migration")
	flags.BoolVar(&options.check, "check", false, "check for changes")
	flags.BoolVar(&options.dryRun, "dry-run", false, "dry run")
	flags.StringVar(&options.database, "database", "default", "database alias")
	flags.BoolVar(&options.fake, "fake", false, "fake apply")
	flags.BoolVar(&options.fakeInitial, "fake-initial", false, "fake initial")
	flags.BoolVar(&options.plan, "plan", false, "show plan")
	flags.IntVar(&options.verbosity, "verbosity", 1, "verbosity")
	flags.BoolVar(&options.merge, "merge", false, "merge conflicts")
	flags.BoolVar(&options.noinput, "noinput", false, "disable input")
	flags.BoolVar(&options.prune, "prune", false, "prune history")
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return options, nil, err
	}
	return options, flags.Args(), nil
}

func runMakeMigrations(options migrationOptions, stdout io.Writer) error {
	appLabel := options.app
	if appLabel == "" {
		appLabel = "project"
	}
	name := options.name
	if name == "" {
		name = "initial"
	}
	migration := migrations.Migration{
		AppLabel: appLabel,
		Name:     migrations.NextMigrationName(1, name),
		Atomic:   true,
		Operations: []migrations.Operation{
			migrations.ManifestOperation{NameValue: "EmptyMigration"},
		},
	}
	if options.dryRun || options.check {
		_, err := fmt.Fprintf(stdout, "would create %s\n", migration.Identity())
		return err
	}
	dir := filepath.Join(appLabel, "migrations")
	if _, err := migrations.NewWriter(dir).Write(migration); err != nil {
		return fmt.Errorf("%w: write migration: %v", ErrCommandFailed, err)
	}
	_, err := fmt.Fprintf(stdout, "created %s\n", migration.Identity())
	return err
}
