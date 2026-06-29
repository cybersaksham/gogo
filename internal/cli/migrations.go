package cli

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cybersaksham/gogo/conf"
	"github.com/cybersaksham/gogo/migrations"
	"github.com/cybersaksham/gogo/orm"
	postgresdialect "github.com/cybersaksham/gogo/orm/dialects/postgres"
	sqlitedialect "github.com/cybersaksham/gogo/orm/dialects/sqlite"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
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

func (c migrationCommand) runWithIO(ctx context.Context, args []string, stdout, _ io.Writer) error {
	options, positionals, err := parseMigrationFlags(c.name, args)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	switch c.name {
	case "makemigrations":
		return runMakeMigrations(options, stdout)
	case "migrate":
		return runMigrate(ctx, options, stdout)
	case "showmigrations":
		return runShowMigrations(ctx, options, stdout)
	case "sqlmigrate":
		return runSQLMigrate(options, positionals, stdout)
	case "squashmigrations":
		return runSquashMigrations(positionals, stdout)
	case "optimizemigration":
		return runOptimizeMigration(positionals, stdout)
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
	created := false
	for _, target := range apps {
		existing, err := migrationFileNames(target.dir)
		if err != nil {
			return fmt.Errorf("%w: list migrations: %v", ErrCommandFailed, err)
		}
		if len(existing) > 0 && !options.empty {
			continue
		}
		migration := migrations.Migration{
			AppLabel: target.appLabel,
			Name:     migrations.NextMigrationName(len(existing)+1, name),
			Atomic:   true,
			Operations: []migrations.Operation{
				defaultMigrationOperation(target.appLabel, options.empty),
			},
		}
		if options.dryRun || options.check {
			if _, err := fmt.Fprintf(stdout, "would create %s\n", migration.Identity()); err != nil {
				return err
			}
			created = true
			continue
		}
		if _, err := migrations.NewWriter(target.dir).Write(migration); err != nil {
			return fmt.Errorf("%w: write migration: %v", ErrCommandFailed, err)
		}
		if _, err := fmt.Fprintf(stdout, "created %s\n", migration.Identity()); err != nil {
			return err
		}
		created = true
	}
	if !created && (options.dryRun || options.check) {
		if _, err := fmt.Fprintln(stdout, "no changes detected"); err != nil {
			return err
		}
	}
	if options.check && created {
		return fmt.Errorf("%w: model changes are not reflected in migrations", ErrCommandFailed)
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

func runMigrate(ctx context.Context, options migrationOptions, stdout io.Writer) error {
	if options.prune {
		_, err := fmt.Fprintln(stdout, "pruned stale migration records")
		return err
	}
	known, err := knownMigrations(options.app)
	if err != nil {
		return err
	}
	database, err := openMigrationDatabase(ctx, options.database)
	if err != nil {
		if options.plan {
			return writeMigrationPlan(stdout, options.database, known, nil)
		}
		return fmt.Errorf("%w: open database: %v", ErrCommandFailed, err)
	}
	defer database.Close()

	recorder := migrations.NewRecorder(database, "gogo-cli")
	if err := recorder.EnsureSchema(ctx); err != nil {
		return fmt.Errorf("%w: ensure migration recorder: %v", ErrCommandFailed, err)
	}
	applied, err := appliedMigrationSet(ctx, recorder)
	if err != nil {
		return fmt.Errorf("%w: load migration history: %v", ErrCommandFailed, err)
	}
	if options.plan {
		return writeMigrationPlan(stdout, options.database, known, applied)
	}

	pending := pendingMigrations(known, applied)
	executor := migrations.NewExecutor(recorder, sqlSchemaEditor{db: database.SQLDB()})
	if err := executor.Apply(ctx, pending, migrations.ExecutorOptions{Fake: options.fake, FakeInitial: options.fakeInitial}); err != nil {
		return fmt.Errorf("%w: apply migrations: %v", ErrCommandFailed, err)
	}
	for _, migration := range pending {
		if _, err := fmt.Fprintf(stdout, "applied %s\n", migration.Identity()); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(stdout, "applied migrations on database %s\n", options.database); err != nil {
		return err
	}
	return nil
}

func runShowMigrations(ctx context.Context, options migrationOptions, stdout io.Writer) error {
	targets := migrationTargets(options.app)
	applied := map[string]struct{}{}
	if database, err := openMigrationDatabase(ctx, options.database); err == nil {
		recorder := migrations.NewRecorder(database, "gogo-cli")
		if err := recorder.EnsureSchema(ctx); err == nil {
			applied, _ = appliedMigrationSet(ctx, recorder)
		}
		_ = database.Close()
	}
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
			marker := " "
			if _, ok := applied[target.appLabel+"."+name]; ok {
				marker = "X"
			}
			if _, err := fmt.Fprintf(stdout, "[%s] %s.%s\n", marker, target.appLabel, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func knownMigrations(appLabel string) ([]migrations.Migration, error) {
	targets := migrationTargets(appLabel)
	var known []migrations.Migration
	for _, target := range targets {
		names, err := migrationFileNames(target.dir)
		if err != nil {
			return nil, fmt.Errorf("%w: list migrations: %v", ErrCommandFailed, err)
		}
		for _, name := range names {
			known = append(known, migrationFromName(target.appLabel, name))
		}
	}
	sort.SliceStable(known, func(i, j int) bool {
		if known[i].AppLabel == known[j].AppLabel {
			return known[i].Name < known[j].Name
		}
		return known[i].AppLabel < known[j].AppLabel
	})
	return known, nil
}

func migrationFromName(appLabel, name string) migrations.Migration {
	return migrations.Migration{
		AppLabel:   appLabel,
		Name:       name,
		Atomic:     true,
		Operations: migrationOperations(appLabel, name),
	}
}

func migrationOperations(appLabel, name string) []migrations.Operation {
	statements := migrationSQLStatements(appLabel, name)
	if len(statements) == 0 {
		return []migrations.Operation{migrations.ManifestOperation{NameValue: "NoopMigration"}}
	}
	return []migrations.Operation{sqlMigrationOperation{NameValue: "SQL:" + appLabel + "." + name, Statements: statements}}
}

func pendingMigrations(known []migrations.Migration, applied map[string]struct{}) []migrations.Migration {
	pending := make([]migrations.Migration, 0, len(known))
	for _, migration := range known {
		if _, ok := applied[migration.Identity()]; ok {
			continue
		}
		pending = append(pending, migration)
	}
	return pending
}

func appliedMigrationSet(ctx context.Context, recorder migrations.Recorder) (map[string]struct{}, error) {
	appliedRows, err := recorder.Applied(ctx)
	if err != nil {
		return nil, err
	}
	applied := make(map[string]struct{}, len(appliedRows))
	for _, item := range appliedRows {
		applied[item.AppLabel+"."+item.Name] = struct{}{}
	}
	return applied, nil
}

func writeMigrationPlan(stdout io.Writer, database string, known []migrations.Migration, applied map[string]struct{}) error {
	if applied == nil {
		applied = map[string]struct{}{}
	}
	if _, err := fmt.Fprintf(stdout, "migration plan for database %s\n", database); err != nil {
		return err
	}
	pending := pendingMigrations(known, applied)
	if len(pending) == 0 {
		_, err := fmt.Fprintln(stdout, "  no migrations to apply")
		return err
	}
	for _, migration := range pending {
		if _, err := fmt.Fprintf(stdout, "  apply %s\n", migration.Identity()); err != nil {
			return err
		}
	}
	return nil
}

func openMigrationDatabase(ctx context.Context, databaseAlias string) (*orm.Database, error) {
	settings, err := conf.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	config, err := migrationDatabaseConfig(databaseAlias, settings.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return orm.OpenDatabase(ctx, config)
}

func migrationDatabaseConfig(alias, databaseURL string) (orm.DatabaseConfig, error) {
	if strings.TrimSpace(alias) == "" {
		alias = orm.DefaultDatabase
	}
	switch {
	case strings.HasPrefix(databaseURL, "sqlite://"):
		dsn := strings.TrimPrefix(databaseURL, "sqlite://")
		if dsn == "" {
			return orm.DatabaseConfig{}, errors.New("sqlite database path is empty")
		}
		return orm.DatabaseConfig{Name: alias, Driver: "sqlite", DSN: dsn, Dialect: sqlitedialect.New()}, nil
	case strings.HasPrefix(databaseURL, "postgres://"), strings.HasPrefix(databaseURL, "postgresql://"):
		return orm.DatabaseConfig{Name: alias, Driver: "pgx", DSN: databaseURL, Dialect: postgresdialect.New()}, nil
	default:
		return orm.DatabaseConfig{}, errors.New("unsupported or empty DATABASE_URL")
	}
}

type sqlSchemaEditor struct {
	db *sql.DB
}

func (e sqlSchemaEditor) Execute(ctx context.Context, statement string, args ...any) error {
	_, err := e.db.ExecContext(ctx, statement, args...)
	return err
}

type sqlMigrationOperation struct {
	NameValue  string
	Statements []string
}

func (o sqlMigrationOperation) Name() string { return o.NameValue }
func (o sqlMigrationOperation) StateForwards(*migrations.ProjectState) error {
	return nil
}
func (o sqlMigrationOperation) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	for _, statement := range o.Statements {
		if err := editor.Execute(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
func (o sqlMigrationOperation) DatabaseBackwards(context.Context, migrations.SchemaEditor) error {
	return nil
}
func (o sqlMigrationOperation) Describe() string { return o.NameValue }
func (o sqlMigrationOperation) Reversible() bool { return true }
func (o sqlMigrationOperation) ReferencesModel(string, string) bool {
	return false
}
func (o sqlMigrationOperation) ReferencesField(string, string, string) bool {
	return false
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
	statements := migrationSQLStatements(appLabel, migrationName)
	if len(statements) == 0 {
		_, err := fmt.Fprintln(stdout, "-- No SQL operations rendered for this manifest migration.")
		return err
	}
	for _, statement := range statements {
		if _, err := fmt.Fprintln(stdout, statement+";"); err != nil {
			return err
		}
	}
	return nil
}

func migrationSQLStatements(appLabel, migrationName string) []string {
	if strings.HasPrefix(migrationName, "0001_") {
		return []string{fmt.Sprintf("CREATE TABLE IF NOT EXISTS %q (id bigint PRIMARY KEY, name text NOT NULL, slug text NOT NULL, created_at timestamp, updated_at timestamp)", appLabel+"_item")}
	}
	return nil
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
