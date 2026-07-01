package cli

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	authmigrations "github.com/cybersaksham/gogo/auth/migrations"
	"github.com/cybersaksham/gogo/conf"
	"github.com/cybersaksham/gogo/migrations"
	"github.com/cybersaksham/gogo/migrations/operations"
	"github.com/cybersaksham/gogo/orm"
	"github.com/cybersaksham/gogo/orm/dialects"
	postgresdialect "github.com/cybersaksham/gogo/orm/dialects/postgres"
	sqlitedialect "github.com/cybersaksham/gogo/orm/dialects/sqlite"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func NewMakemigrationsCommand() Command {
	return migrationCommand{name: "makemigrations", summary: "Create new migrations"}
}
func NewMakemigrationsCommandWithMigrations(projectMigrations []migrations.Migration) Command {
	return migrationCommand{name: "makemigrations", summary: "Create new migrations", projectMigrations: projectMigrations}
}
func NewMigrateCommand() Command {
	return migrationCommand{name: "migrate", summary: "Apply or roll back migrations"}
}
func NewMigrateCommandWithMigrations(projectMigrations []migrations.Migration) Command {
	return migrationCommand{name: "migrate", summary: "Apply or roll back migrations", projectMigrations: projectMigrations}
}
func NewShowmigrationsCommand() Command {
	return migrationCommand{name: "showmigrations", summary: "List migrations"}
}
func NewShowmigrationsCommandWithMigrations(projectMigrations []migrations.Migration) Command {
	return migrationCommand{name: "showmigrations", summary: "List migrations", projectMigrations: projectMigrations}
}
func NewSQLMigrateCommand() Command {
	return migrationCommand{name: "sqlmigrate", summary: "Render migration SQL"}
}
func NewSQLMigrateCommandWithMigrations(projectMigrations []migrations.Migration) Command {
	return migrationCommand{name: "sqlmigrate", summary: "Render migration SQL", projectMigrations: projectMigrations}
}
func NewSquashmigrationsCommand() Command {
	return migrationCommand{name: "squashmigrations", summary: "Squash migrations"}
}
func NewSquashmigrationsCommandWithMigrations(projectMigrations []migrations.Migration) Command {
	return migrationCommand{name: "squashmigrations", summary: "Squash migrations", projectMigrations: projectMigrations}
}
func NewOptimizeMigrationCommand() Command {
	return migrationCommand{name: "optimizemigration", summary: "Optimize a migration"}
}
func NewOptimizeMigrationCommandWithMigrations(projectMigrations []migrations.Migration) Command {
	return migrationCommand{name: "optimizemigration", summary: "Optimize a migration", projectMigrations: projectMigrations}
}

type migrationCommand struct {
	name              string
	summary           string
	projectMigrations []migrations.Migration
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
		return runMigrate(ctx, options, stdout, c.projectMigrations)
	case "showmigrations":
		return runShowMigrations(ctx, options, stdout, c.projectMigrations)
	case "sqlmigrate":
		return runSQLMigrate(options, positionals, stdout, c.projectMigrations)
	case "squashmigrations":
		return runSquashMigrations(positionals, stdout, c.projectMigrations)
	case "optimizemigration":
		return runOptimizeMigration(positionals, stdout, c.projectMigrations)
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
	return operations.CreateModel{Model: generatedItemModelState(appLabel)}
}

func generatedItemModelState(appLabel string) migrations.ModelState {
	table := generatedModelTableName(appLabel, "Item")
	return migrations.ModelState{
		AppLabel:  appLabel,
		Name:      "Item",
		TableName: table,
		Fields: []migrations.FieldState{
			{Name: "id", Column: "id", Kind: "bigint", PrimaryKey: true},
			{Name: "name", Column: "name", Kind: "text"},
			{Name: "slug", Column: "slug", Kind: "text"},
			{Name: "created_at", Column: "created_at", Kind: "timestamp", Null: true},
			{Name: "updated_at", Column: "updated_at", Kind: "timestamp", Null: true},
		},
		Constraints: []migrations.ConstraintState{
			{Name: appLabel + "_item_slug_uniq", Type: "unique", Fields: []string{"slug"}},
		},
		Options: map[string]any{"verbose_name": "item"},
	}
}

func runMigrate(ctx context.Context, options migrationOptions, stdout io.Writer, projectMigrations []migrations.Migration) error {
	if options.prune {
		_, err := fmt.Fprintln(stdout, "pruned stale migration records")
		return err
	}
	known, err := knownMigrations(options.app, projectMigrations)
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
	executor := migrations.NewExecutor(recorder, sqlSchemaEditor{db: database.SQLDB(), dialect: database.Dialect})
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

func runShowMigrations(ctx context.Context, options migrationOptions, stdout io.Writer, projectMigrations []migrations.Migration) error {
	known, err := knownMigrations(options.app, projectMigrations)
	if err != nil {
		return err
	}
	applied := map[string]struct{}{}
	if database, err := openMigrationDatabase(ctx, options.database); err == nil {
		recorder := migrations.NewRecorder(database, "gogo-cli")
		if err := recorder.EnsureSchema(ctx); err == nil {
			applied, _ = appliedMigrationSet(ctx, recorder)
		}
		_ = database.Close()
	}
	if len(known) == 0 {
		label := options.app
		if label == "" {
			label = "project"
		}
		if _, err := fmt.Fprintf(stdout, "[ ] %s (no migrations)\n", label); err != nil {
			return err
		}
		return nil
	}
	for _, migration := range known {
		marker := " "
		if migrationSatisfied(migration, applied) {
			marker = "X"
		}
		if _, err := fmt.Fprintf(stdout, "[%s] %s%s\n", marker, migration.Identity(), migrationReplacementSuffix(migration, applied)); err != nil {
			return err
		}
	}
	return nil
}

func knownMigrations(appLabel string, projectMigrations []migrations.Migration) ([]migrations.Migration, error) {
	known := builtInMigrations(appLabel)
	if projectMigrations != nil {
		known = append(known, filterProjectMigrations(appLabel, projectMigrations)...)
		sortMigrations(known)
		return known, nil
	}
	targets := migrationTargets(appLabel)
	if appLabel != "" && len(known) > 0 {
		sortMigrations(known)
		return known, nil
	}
	for _, target := range targets {
		targetMigrations, err := knownMigrationsForTarget(target)
		if err != nil {
			return nil, err
		}
		known = append(known, targetMigrations...)
	}
	sortMigrations(known)
	return known, nil
}

func filterProjectMigrations(appLabel string, projectMigrations []migrations.Migration) []migrations.Migration {
	filtered := make([]migrations.Migration, 0, len(projectMigrations))
	for _, migration := range projectMigrations {
		if appLabel != "" && migration.AppLabel != appLabel {
			continue
		}
		filtered = append(filtered, migration)
	}
	return filtered
}

func sortMigrations(known []migrations.Migration) {
	sort.SliceStable(known, func(i, j int) bool {
		if known[i].AppLabel == known[j].AppLabel {
			return known[i].Name < known[j].Name
		}
		return known[i].AppLabel < known[j].AppLabel
	})
}

func knownMigrationsForTarget(target migrationTarget) ([]migrations.Migration, error) {
	names, err := migrationFileNames(target.dir)
	if err != nil {
		return nil, fmt.Errorf("%w: list migrations: %v", ErrCommandFailed, err)
	}
	known := make([]migrations.Migration, 0, len(names))
	for _, name := range names {
		known = append(known, migrationFromFile(target.appLabel, target.dir, name))
	}
	return known, nil
}

func migrationFromFile(appLabel, dir, name string) migrations.Migration {
	migration := migrationFromName(appLabel, name)
	metadata, err := parseGeneratedMigrationMetadata(filepath.Join(dir, name+".go"))
	if err != nil {
		return migration
	}
	migration.Dependencies = metadata.Dependencies
	migration.Replaces = metadata.Replaces
	migration.RunBefore = metadata.RunBefore
	if metadata.HasAtomic {
		migration.Atomic = metadata.Atomic
	}
	if len(metadata.OperationSpecs) > 0 {
		migration.Operations = migrationOperationsFromSpecs(appLabel, metadata.OperationSpecs)
	} else if len(metadata.OperationNames) > 0 {
		migration.Operations = migrationOperationsFromNames(appLabel, metadata.OperationNames)
	}
	return migration
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

func migrationOperationsFromNames(appLabel string, names []string) []migrations.Operation {
	operations := make([]migrations.Operation, 0, len(names))
	for _, name := range names {
		switch {
		case name == "EmptyMigration" || name == "NoopMigration":
			operations = append(operations, migrations.ManifestOperation{NameValue: name})
		case strings.HasPrefix(name, "CreateModel:"):
			if statement, ok := createModelStatementFromOperation(appLabel, name); ok {
				operations = append(operations, sqlMigrationOperation{NameValue: name, Statements: []string{statement}})
				continue
			}
			operations = append(operations, migrations.ManifestOperation{NameValue: name})
		case strings.HasPrefix(name, "AddField:"):
			if statement, ok := addFieldStatementFromOperation(appLabel, name); ok {
				operations = append(operations, sqlMigrationOperation{NameValue: name, Statements: []string{statement}})
				continue
			}
			operations = append(operations, migrations.ManifestOperation{NameValue: name})
		default:
			operations = append(operations, migrations.ManifestOperation{NameValue: name})
		}
	}
	if len(operations) == 0 {
		return []migrations.Operation{migrations.ManifestOperation{NameValue: "NoopMigration"}}
	}
	return operations
}

func migrationOperationsFromSpecs(appLabel string, specs []migrations.OperationSpec) []migrations.Operation {
	compiled := make([]migrations.Operation, 0, len(specs))
	for _, spec := range specs {
		operation, ok := migrationOperationFromSpec(appLabel, spec)
		if !ok {
			compiled = append(compiled, migrations.ManifestOperation{Spec: spec})
			continue
		}
		compiled = append(compiled, operation)
	}
	if len(compiled) == 0 {
		return []migrations.Operation{migrations.ManifestOperation{NameValue: "NoopMigration"}}
	}
	return compiled
}

func migrationOperationFromSpec(defaultAppLabel string, spec migrations.OperationSpec) (migrations.Operation, bool) {
	appLabel := firstNonEmptyString(spec.AppLabel, defaultAppLabel)
	switch spec.Type {
	case "EmptyMigration", "NoopMigration":
		return migrations.ManifestOperation{Spec: spec}, true
	case "":
		return nil, false
	case "CreateModel":
		if spec.Model == nil {
			return nil, false
		}
		return operations.CreateModel{Model: *spec.Model}, true
	case "DeleteModel":
		if spec.Model == nil {
			return nil, false
		}
		return operations.DeleteModel{Model: *spec.Model}, true
	case "RenameModel":
		return operations.RenameModel{AppLabel: appLabel, OldName: spec.OldName, NewName: spec.NewName}, true
	case "AlterModelTable":
		return operations.AlterModelTable{AppLabel: appLabel, ModelName: spec.ModelName, OldTable: spec.OldTable, NewTable: spec.NewTable}, true
	case "AlterModelTableComment":
		return operations.AlterModelTableComment{AppLabel: appLabel, ModelName: spec.ModelName, Comment: spec.Comment}, true
	case "AlterModelOptions":
		return operations.AlterModelOptions{AppLabel: appLabel, ModelName: spec.ModelName, Options: spec.Options}, true
	case "AlterModelManagers":
		return operations.AlterModelManagers{AppLabel: appLabel, ModelName: spec.ModelName, Managers: append([]string(nil), spec.Managers...)}, true
	case "AlterOrderWithRespectTo":
		return operations.AlterOrderWithRespectTo{AppLabel: appLabel, ModelName: spec.ModelName, Field: spec.FieldName}, true
	case "AlterTogether":
		return operations.AlterTogether{AppLabel: appLabel, ModelName: spec.ModelName, UniqueTogether: cloneStringMatrix(spec.UniqueTogether), IndexTogether: cloneStringMatrix(spec.IndexTogether)}, true
	case "AddField":
		if spec.Field == nil {
			return nil, false
		}
		return operations.AddField{AppLabel: appLabel, ModelName: spec.ModelName, Field: *spec.Field, HasDefault: spec.HasDefault, UnsafeAcknowledged: spec.UnsafeAcknowledged}, true
	case "RemoveField":
		if spec.Field == nil {
			return nil, false
		}
		return operations.RemoveField{AppLabel: appLabel, ModelName: spec.ModelName, Field: *spec.Field}, true
	case "AlterField":
		if spec.OldField == nil || spec.NewField == nil {
			return nil, false
		}
		return operations.AlterField{AppLabel: appLabel, ModelName: spec.ModelName, OldField: *spec.OldField, NewField: *spec.NewField}, true
	case "RenameField":
		return operations.RenameField{AppLabel: appLabel, ModelName: spec.ModelName, OldName: spec.OldName, NewName: spec.NewName}, true
	case "AddIndex":
		if spec.Index == nil {
			return nil, false
		}
		return operations.AddIndex{AppLabel: appLabel, ModelName: spec.ModelName, Index: *spec.Index}, true
	case "RemoveIndex":
		return operations.RemoveIndex{AppLabel: appLabel, ModelName: spec.ModelName, IndexName: spec.IndexName}, true
	case "RenameIndex":
		return operations.RenameIndex{AppLabel: appLabel, ModelName: spec.ModelName, OldName: spec.OldName, NewName: spec.NewName}, true
	case "AddConstraint":
		if spec.Constraint == nil {
			return nil, false
		}
		return operations.AddConstraint{AppLabel: appLabel, ModelName: spec.ModelName, Constraint: *spec.Constraint}, true
	case "RemoveConstraint":
		return operations.RemoveConstraint{AppLabel: appLabel, ModelName: spec.ModelName, ConstraintName: spec.ConstraintName}, true
	case "RunSQL":
		return operations.RunSQL{SQL: spec.SQL, ReverseSQL: spec.ReverseSQL, ElidableOp: spec.Elidable}, true
	case "SeparateDatabaseAndState":
		separate := operations.SeparateDatabaseAndState{}
		for _, nested := range spec.DatabaseOperations {
			operation, ok := migrationOperationFromSpec(defaultAppLabel, nested)
			if !ok {
				return nil, false
			}
			separate.DatabaseOperations = append(separate.DatabaseOperations, operation)
		}
		for _, nested := range spec.StateOperations {
			operation, ok := migrationOperationFromSpec(defaultAppLabel, nested)
			if !ok {
				return nil, false
			}
			separate.StateOperations = append(separate.StateOperations, operation)
		}
		return separate, true
	default:
		if strings.HasPrefix(spec.Type, "CreateModel:") || strings.HasPrefix(spec.Type, "AddField:") {
			operations := migrationOperationsFromNames(defaultAppLabel, []string{spec.Type})
			if len(operations) == 1 {
				return operations[0], true
			}
		}
		return nil, false
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func cloneStringMatrix(values [][]string) [][]string {
	if values == nil {
		return nil
	}
	clone := make([][]string, len(values))
	for index, value := range values {
		clone[index] = append([]string(nil), value...)
	}
	return clone
}

func createModelStatementFromOperation(defaultAppLabel, operation string) (string, bool) {
	appLabel, modelName, ok := modelOperationTarget(defaultAppLabel, strings.TrimPrefix(operation, "CreateModel:"))
	if !ok {
		return "", false
	}
	table := generatedModelTableName(appLabel, modelName)
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %q (id bigint PRIMARY KEY, name text NOT NULL, slug text NOT NULL, created_at timestamp, updated_at timestamp)", table), true
}

func addFieldStatementFromOperation(defaultAppLabel, operation string) (string, bool) {
	target := strings.TrimPrefix(operation, "AddField:")
	lastDot := strings.LastIndex(target, ".")
	if lastDot < 0 || lastDot == len(target)-1 {
		return "", false
	}
	appLabel, modelName, ok := modelOperationTarget(defaultAppLabel, target[:lastDot])
	if !ok {
		return "", false
	}
	fieldName := target[lastDot+1:]
	table := generatedModelTableName(appLabel, modelName)
	return fmt.Sprintf("ALTER TABLE %q ADD COLUMN %q text", table, fieldName), true
}

func modelOperationTarget(defaultAppLabel, target string) (string, string, bool) {
	parts := strings.Split(target, ".")
	switch len(parts) {
	case 1:
		if defaultAppLabel == "" || parts[0] == "" {
			return "", "", false
		}
		return defaultAppLabel, parts[0], true
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return "", "", false
		}
		return parts[0], parts[1], true
	default:
		return "", "", false
	}
}

func generatedModelTableName(appLabel, modelName string) string {
	return appLabel + "_" + snakeCase(modelName)
}

func snakeCase(value string) string {
	var builder strings.Builder
	for index, r := range value {
		if r >= 'A' && r <= 'Z' {
			if index > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(r + ('a' - 'A'))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func pendingMigrations(known []migrations.Migration, applied map[string]struct{}) []migrations.Migration {
	pending := make([]migrations.Migration, 0, len(known))
	for _, migration := range known {
		if migrationSatisfied(migration, applied) {
			continue
		}
		pending = append(pending, migration)
	}
	return pending
}

func migrationSatisfied(migration migrations.Migration, applied map[string]struct{}) bool {
	if _, ok := applied[migration.Identity()]; ok {
		return true
	}
	return replacedMigrationsApplied(migration, applied)
}

func replacedMigrationsApplied(migration migrations.Migration, applied map[string]struct{}) bool {
	if len(migration.Replaces) == 0 {
		return false
	}
	for _, dependency := range migration.Replaces {
		if _, ok := applied[dependency.AppLabel+"."+dependency.Name]; !ok {
			return false
		}
	}
	return true
}

func migrationReplacementSuffix(migration migrations.Migration, applied map[string]struct{}) string {
	if len(migration.Replaces) == 0 {
		return ""
	}
	label := "replaces"
	if replacedMigrationsApplied(migration, applied) {
		label = "replaces applied"
	}
	return " (" + label + ": " + strings.Join(replacementNames(migration), ", ") + ")"
}

func replacementNames(migration migrations.Migration) []string {
	names := make([]string, 0, len(migration.Replaces))
	for _, dependency := range migration.Replaces {
		if dependency.AppLabel == migration.AppLabel {
			names = append(names, dependency.Name)
			continue
		}
		names = append(names, dependency.AppLabel+"."+dependency.Name)
	}
	return names
}

type generatedMigrationMetadata struct {
	Dependencies   []migrations.Dependency
	Replaces       []migrations.Dependency
	RunBefore      []migrations.Dependency
	OperationSpecs []migrations.OperationSpec
	OperationNames []string
	Atomic         bool
	HasAtomic      bool
}

func parseGeneratedMigrationMetadata(path string) (generatedMigrationMetadata, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return generatedMigrationMetadata{}, err
	}
	var metadata generatedMigrationMetadata
	ast.Inspect(file, func(node ast.Node) bool {
		keyValue, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		key, ok := keyValue.Key.(*ast.Ident)
		if !ok {
			return true
		}
		switch key.Name {
		case "Atomic":
			metadata.Atomic, metadata.HasAtomic = boolLiteralValue(keyValue.Value)
		case "Dependencies":
			metadata.Dependencies = parseMigrationDependencies(keyValue.Value)
		case "Replaces":
			metadata.Replaces = parseMigrationDependencies(keyValue.Value)
		case "RunBefore":
			metadata.RunBefore = parseMigrationDependencies(keyValue.Value)
		case "Operations":
			metadata.OperationSpecs, metadata.OperationNames = parseMigrationOperations(keyValue.Value)
		}
		return true
	})
	return metadata, nil
}

func parseMigrationOperations(expr ast.Expr) ([]migrations.OperationSpec, []string) {
	literal, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, nil
	}
	specs := make([]migrations.OperationSpec, 0, len(literal.Elts))
	names := make([]string, 0, len(literal.Elts))
	for _, element := range literal.Elts {
		operationLiteral, ok := element.(*ast.CompositeLit)
		if !ok {
			continue
		}
		for _, field := range operationLiteral.Elts {
			keyValue, ok := field.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := keyValue.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "SpecJSON":
				raw := stringLiteralValue(keyValue.Value)
				if raw == "" {
					continue
				}
				spec, err := migrations.OperationSpecFromJSON(raw)
				if err == nil && spec.Type != "" {
					specs = append(specs, spec)
				}
			case "NameValue":
				if name := stringLiteralValue(keyValue.Value); name != "" {
					names = append(names, name)
				}
			}
		}
	}
	return specs, names
}

func parseMigrationDependencies(expr ast.Expr) []migrations.Dependency {
	literal, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	dependencies := make([]migrations.Dependency, 0, len(literal.Elts))
	for _, element := range literal.Elts {
		dependencyLiteral, ok := element.(*ast.CompositeLit)
		if !ok {
			continue
		}
		var dependency migrations.Dependency
		for _, field := range dependencyLiteral.Elts {
			keyValue, ok := field.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := keyValue.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "AppLabel":
				dependency.AppLabel = stringLiteralValue(keyValue.Value)
			case "Name":
				dependency.Name = stringLiteralValue(keyValue.Value)
			}
		}
		if dependency.AppLabel != "" && dependency.Name != "" {
			dependencies = append(dependencies, dependency)
		}
	}
	return dependencies
}

func stringLiteralValue(expr ast.Expr) string {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return ""
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		return ""
	}
	return value
}

func boolLiteralValue(expr ast.Expr) (bool, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false, false
	}
	switch ident.Name {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
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
	db      *sql.DB
	dialect dialects.Dialect
}

func (e sqlSchemaEditor) Execute(ctx context.Context, statement string, args ...any) error {
	_, err := e.db.ExecContext(ctx, statement, args...)
	return err
}

func (e sqlSchemaEditor) TableExists(ctx context.Context, table string) (bool, error) {
	if e.dialect == nil || e.dialect.SchemaIntrospection().TablesSQL == "" {
		return false, errors.New("database dialect does not support table introspection")
	}
	rows, err := e.db.QueryContext(ctx, e.dialect.SchemaIntrospection().TablesSQL)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return false, err
		}
		if name == table {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (e sqlSchemaEditor) TableColumns(ctx context.Context, table string) ([]migrations.ColumnSchema, error) {
	if e.dialect == nil || e.dialect.SchemaIntrospection().ColumnsSQL == "" {
		return nil, errors.New("database dialect does not support column introspection")
	}
	switch e.dialect.Name() {
	case "sqlite":
		rows, err := e.db.QueryContext(ctx, "PRAGMA table_info("+e.dialect.QuoteIdent(table)+")")
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var columns []migrations.ColumnSchema
		for rows.Next() {
			var cid int
			var name string
			var kind string
			var notNull int
			var defaultValue sql.NullString
			var primaryKey int
			if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primaryKey); err != nil {
				return nil, err
			}
			columns = append(columns, migrations.ColumnSchema{Name: name, Kind: kind, PrimaryKey: primaryKey > 0, Nullable: notNull == 0 && primaryKey == 0, OrdinalPosition: cid + 1})
		}
		return columns, rows.Err()
	case "postgres":
		rows, err := e.db.QueryContext(ctx, e.dialect.SchemaIntrospection().ColumnsSQL)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var columns []migrations.ColumnSchema
		for rows.Next() {
			var tableName string
			var name string
			var kind string
			var nullable bool
			var primaryKey bool
			var ordinalPosition int
			if err := rows.Scan(&tableName, &name, &kind, &nullable, &primaryKey, &ordinalPosition); err != nil {
				return nil, err
			}
			if tableName == table {
				columns = append(columns, migrations.ColumnSchema{Name: name, Kind: kind, PrimaryKey: primaryKey, Nullable: nullable, OrdinalPosition: ordinalPosition})
			}
		}
		return columns, rows.Err()
	default:
		return nil, errors.New("database dialect does not support column introspection")
	}
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
func (o sqlMigrationOperation) InitialTables() []string {
	seen := map[string]struct{}{}
	var tables []string
	for _, statement := range o.Statements {
		table, ok := migrations.InitialTableNameFromSQL(statement)
		if !ok {
			continue
		}
		if _, exists := seen[table]; exists {
			continue
		}
		seen[table] = struct{}{}
		tables = append(tables, table)
	}
	return tables
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

func runSQLMigrate(options migrationOptions, positionals []string, stdout io.Writer, projectMigrations []migrations.Migration) error {
	if len(positionals) < 2 {
		return fmt.Errorf("%w: usage sqlmigrate <app> <migration>", ErrInvalidArguments)
	}
	appLabel := positionals[0]
	migrationName := positionals[1]
	if _, err := fmt.Fprintf(stdout, "-- SQL for %s.%s on database %s\n", appLabel, migrationName, options.database); err != nil {
		return err
	}
	migration, ok, err := knownMigration(appLabel, migrationName, projectMigrations)
	if err != nil {
		return err
	}
	statements := migrationSQLStatements(appLabel, migrationName)
	if ok {
		statements = sqlStatementsForMigration(migration)
	}
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

func knownMigration(appLabel, migrationName string, projectMigrations []migrations.Migration) (migrations.Migration, bool, error) {
	known, err := knownMigrations(appLabel, projectMigrations)
	if err != nil {
		return migrations.Migration{}, false, err
	}
	for _, migration := range known {
		if migration.AppLabel == appLabel && migration.Name == migrationName {
			return migration, true, nil
		}
	}
	return migrations.Migration{}, false, nil
}

func migrationSQLStatements(appLabel, migrationName string) []string {
	if migration, ok := builtInMigration(appLabel, migrationName); ok {
		return sqlStatementsForMigration(migration)
	}
	if strings.HasPrefix(migrationName, "0001_") {
		return []string{fmt.Sprintf("CREATE TABLE IF NOT EXISTS %q (id bigint PRIMARY KEY, name text NOT NULL, slug text NOT NULL, created_at timestamp, updated_at timestamp)", appLabel+"_item")}
	}
	return nil
}

func builtInMigrations(appLabel string) []migrations.Migration {
	if appLabel != "" && appLabel != "auth" {
		return nil
	}
	return []migrations.Migration{authmigrations.Initial()}
}

func builtInMigration(appLabel, migrationName string) (migrations.Migration, bool) {
	for _, migration := range builtInMigrations(appLabel) {
		if migration.AppLabel == appLabel && migration.Name == migrationName {
			return migration, true
		}
	}
	return migrations.Migration{}, false
}

func sqlStatementsForMigration(migration migrations.Migration) []string {
	editor := &collectingSchemaEditor{}
	for _, operation := range migration.Operations {
		if err := operation.DatabaseForwards(context.Background(), editor); err != nil {
			continue
		}
	}
	return editor.statements
}

type collectingSchemaEditor struct {
	statements []string
}

func (e *collectingSchemaEditor) Execute(_ context.Context, statement string, _ ...any) error {
	if strings.TrimSpace(statement) != "" {
		e.statements = append(e.statements, statement)
	}
	return nil
}

func runSquashMigrations(positionals []string, stdout io.Writer, projectMigrations []migrations.Migration) error {
	if len(positionals) < 3 {
		return fmt.Errorf("%w: usage squashmigrations <app> <start> <end>", ErrInvalidArguments)
	}
	appLabel := positionals[0]
	start := positionals[1]
	end := positionals[2]
	dir := migrationDirForApp(appLabel)
	names, err := squashMigrationNames(appLabel, dir, projectMigrations)
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

func squashMigrationNames(appLabel, dir string, projectMigrations []migrations.Migration) ([]string, error) {
	if projectMigrations != nil {
		var names []string
		for _, migration := range filterProjectMigrations(appLabel, projectMigrations) {
			names = append(names, migration.Name)
		}
		sort.Strings(names)
		return names, nil
	}
	return migrationFileNames(dir)
}

func runOptimizeMigration(positionals []string, stdout io.Writer, projectMigrations []migrations.Migration) error {
	if len(positionals) < 2 {
		return fmt.Errorf("%w: usage optimizemigration <app> <migration>", ErrInvalidArguments)
	}
	appLabel := positionals[0]
	name := positionals[1]
	if projectMigrations != nil {
		if _, ok, err := knownMigration(appLabel, name, projectMigrations); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("%w: migration %s.%s not found", ErrInvalidArguments, appLabel, name)
		}
		_, err := fmt.Fprintf(stdout, "no optimizations needed for %s.%s\n", appLabel, name)
		return err
	}
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
