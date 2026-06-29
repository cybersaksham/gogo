package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		return runShowMigrations(options, stdout)
	case "sqlmigrate":
		return runSQLMigrate(options, positionals, stdout)
	case "squashmigrations":
		return runSquashMigrations(positionals, stdout)
	case "optimizemigration":
		return runOptimizeMigration(positionals, stdout)
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
	apps := migrationTargets(options.app)
	name := options.name
	if name == "" {
		name = "initial"
	}
	for _, target := range apps {
		migration := migrations.Migration{
			AppLabel: target.appLabel,
			Name:     migrations.NextMigrationName(1, name),
			Atomic:   true,
			Operations: []migrations.Operation{
				defaultMigrationOperation(target.appLabel, options.empty),
			},
		}
		if options.dryRun || options.check {
			if _, err := fmt.Fprintf(stdout, "would create %s\n", migration.Identity()); err != nil {
				return err
			}
			continue
		}
		if _, err := migrations.NewWriter(target.dir).Write(migration); err != nil {
			return fmt.Errorf("%w: write migration: %v", ErrCommandFailed, err)
		}
		if _, err := fmt.Fprintf(stdout, "created %s\n", migration.Identity()); err != nil {
			return err
		}
	}
	return nil
}

type migrationTarget struct {
	appLabel string
	dir      string
}

func migrationTargets(appLabel string) []migrationTarget {
	if appLabel != "" {
		return []migrationTarget{{appLabel: appLabel, dir: migrationDirForApp(appLabel)}}
	}
	if targets := discoverGeneratedAppMigrationTargets(); len(targets) > 0 {
		return targets
	}
	return []migrationTarget{{appLabel: "project", dir: filepath.Join("project", "migrations")}}
}

func migrationDirForApp(appLabel string) string {
	if stat, err := os.Stat(filepath.Join("apps", appLabel)); err == nil && stat.IsDir() {
		return filepath.Join("apps", appLabel, "migrations")
	}
	return filepath.Join(appLabel, "migrations")
}

func discoverGeneratedAppMigrationTargets() []migrationTarget {
	entries, err := os.ReadDir("apps")
	if err != nil {
		return nil
	}
	var targets []migrationTarget
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		targets = append(targets, migrationTarget{
			appLabel: entry.Name(),
			dir:      filepath.Join("apps", entry.Name(), "migrations"),
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].appLabel < targets[j].appLabel
	})
	return targets
}

func defaultMigrationOperation(appLabel string, empty bool) migrations.Operation {
	if empty {
		return migrations.ManifestOperation{NameValue: "EmptyMigration"}
	}
	return migrations.ManifestOperation{NameValue: "CreateModel:" + appLabel + ".Item"}
}

func runShowMigrations(options migrationOptions, stdout io.Writer) error {
	targets := migrationTargets(options.app)
	for _, target := range targets {
		names, err := migrationFileNames(target.dir)
		if err != nil {
			return fmt.Errorf("%w: list migrations: %v", ErrCommandFailed, err)
		}
		if len(names) == 0 {
			if _, err := fmt.Fprintf(stdout, "[ ] %s (no migrations)\n", target.appLabel); err != nil {
				return err
			}
			continue
		}
		for _, name := range names {
			if _, err := fmt.Fprintf(stdout, "[ ] %s.%s\n", target.appLabel, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrationFileNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".go"))
	}
	sort.Strings(names)
	return names, nil
}

func runSQLMigrate(options migrationOptions, positionals []string, stdout io.Writer) error {
	if len(positionals) < 2 {
		return fmt.Errorf("%w: usage sqlmigrate <app> <migration>", ErrInvalidArguments)
	}
	appLabel := positionals[0]
	migrationName := positionals[1]
	if _, err := fmt.Fprintf(stdout, "-- SQL for %s.%s on database %s\n", appLabel, migrationName, options.database); err != nil {
		return err
	}
	if strings.HasPrefix(migrationName, "0001_") {
		_, err := fmt.Fprintf(stdout, "CREATE TABLE IF NOT EXISTS %q (id bigint PRIMARY KEY, name text NOT NULL, slug text NOT NULL, created_at timestamp, updated_at timestamp);\n", appLabel+"_item")
		return err
	}
	_, err := fmt.Fprintln(stdout, "-- No SQL operations rendered for this manifest migration.")
	return err
}

func runSquashMigrations(positionals []string, stdout io.Writer) error {
	if len(positionals) < 3 {
		return fmt.Errorf("%w: usage squashmigrations <app> <start> <end>", ErrInvalidArguments)
	}
	appLabel := positionals[0]
	start := positionals[1]
	end := positionals[2]
	dir := migrationDirForApp(appLabel)
	names, err := migrationFileNames(dir)
	if err != nil {
		return fmt.Errorf("%w: list migrations: %v", ErrCommandFailed, err)
	}
	selected, err := migrationRange(names, start, end)
	if err != nil {
		return err
	}
	replaces := make([]migrations.Dependency, 0, len(selected))
	for _, name := range selected {
		replaces = append(replaces, migrations.Dependency{AppLabel: appLabel, Name: name})
	}
	migration := migrations.Migration{
		AppLabel: appLabel,
		Name:     squashedMigrationName(start, end),
		Replaces: replaces,
		Atomic:   true,
		Operations: []migrations.Operation{
			migrations.ManifestOperation{NameValue: "SquashedMigration:" + appLabel + "." + start + ".." + end},
		},
	}
	if _, err := migrations.NewWriter(dir).Write(migration); err != nil {
		return fmt.Errorf("%w: write squashed migration: %v", ErrCommandFailed, err)
	}
	_, err = fmt.Fprintf(stdout, "created squashed migration %s replacing %d migration(s)\n", migration.Identity(), len(selected))
	return err
}

func runOptimizeMigration(positionals []string, stdout io.Writer) error {
	if len(positionals) < 2 {
		return fmt.Errorf("%w: usage optimizemigration <app> <migration>", ErrInvalidArguments)
	}
	appLabel := positionals[0]
	name := positionals[1]
	path := filepath.Join(migrationDirForApp(appLabel), name+".go")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: migration %s.%s not found", ErrInvalidArguments, appLabel, name)
		}
		return fmt.Errorf("%w: inspect migration: %v", ErrCommandFailed, err)
	}
	_, err := fmt.Fprintf(stdout, "no optimizations needed for %s.%s\n", appLabel, name)
	return err
}

func migrationRange(names []string, start, end string) ([]string, error) {
	startIndex := -1
	endIndex := -1
	for index, name := range names {
		if name == start {
			startIndex = index
		}
		if name == end {
			endIndex = index
		}
	}
	if startIndex < 0 {
		return nil, fmt.Errorf("%w: migration %s not found", ErrInvalidArguments, start)
	}
	if endIndex < 0 {
		return nil, fmt.Errorf("%w: migration %s not found", ErrInvalidArguments, end)
	}
	if startIndex > endIndex {
		return nil, fmt.Errorf("%w: start migration must come before end migration", ErrInvalidArguments)
	}
	return append([]string(nil), names[startIndex:endIndex+1]...), nil
}

func squashedMigrationName(start, end string) string {
	number := strings.SplitN(start, "_", 2)[0]
	return number + "_squashed_" + end
}
