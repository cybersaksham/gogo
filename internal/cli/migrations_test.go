package cli

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	authmigrations "github.com/cybersaksham/gogo/auth/migrations"
	"github.com/cybersaksham/gogo/migrations"

	_ "modernc.org/sqlite"
)

func TestMigrationCommandsRunWithFlags(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	migrationsDir := filepath.Join(dir, "blog", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	writeTextFile(t, filepath.Join(migrationsDir, "0001_initial.go"), "package migrations\n")
	writeTextFile(t, filepath.Join(migrationsDir, "0002_post.go"), "package migrations\n")
	root := NewRoot()
	cases := [][]string{
		{"makemigrations", "--app", "blog", "--name", "initial", "--empty", "--dry-run"},
		{"migrate", "--database", "default", "--plan", "--fake"},
		{"showmigrations", "--app", "blog", "--verbosity", "2"},
		{"sqlmigrate", "blog", "0001_initial", "--database", "default"},
		{"squashmigrations", "blog", "0001_initial", "0002_post", "--noinput"},
		{"migrate", "--prune"},
		{"optimizemigration", "blog", "0001_initial"},
	}
	for _, args := range cases {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if err := root.Execute(context.Background(), args, &stdout, &stderr); err != nil {
			t.Fatalf("Execute(%v) error = %v", args, err)
		}
		if stdout.Len() == 0 {
			t.Fatalf("Execute(%v) produced no output", args)
		}
	}
}

func TestMakeMigrationsWritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	root := NewRoot()
	var stdout bytes.Buffer
	if err := root.Execute(context.Background(), []string{"makemigrations", "--app", "blog", "--name", "initial", "--empty"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("makemigrations error = %v", err)
	}
	path := filepath.Join(dir, "blog", "migrations", "0001_initial.go")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected migration file %s: %v", path, err)
	}
	if !strings.Contains(stdout.String(), "created blog.0001_initial") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestMakeMigrationsDiscoversGeneratedAppsFromProjectRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "apps", "blog", "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir blog migrations: %v", err)
	}
	t.Chdir(dir)

	root := NewRoot()
	var stdout bytes.Buffer
	if err := root.Execute(context.Background(), []string{"makemigrations"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("makemigrations error = %v", err)
	}
	path := filepath.Join(dir, "apps", "blog", "migrations", "0001_initial.go")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected generated app migration %s: %v", path, err)
	}
	if !strings.Contains(string(contents), `"blog"`) || !strings.Contains(string(contents), `CreateModel:blog.Item`) {
		t.Fatalf("migration contents did not describe blog item:\n%s", contents)
	}
	if !strings.Contains(stdout.String(), "created blog.0001_initial") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestShowAndSQLMigrateUseGeneratedAppOutput(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	writeTextFile(t, filepath.Join(migrationsDir, "0001_initial.go"), "package migrations\n")
	t.Chdir(dir)

	root := NewRoot()
	var showOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"showmigrations"}, &showOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("showmigrations error = %v", err)
	}
	if !strings.Contains(showOut.String(), "[ ] blog.0001_initial") {
		t.Fatalf("showmigrations stdout = %q", showOut.String())
	}

	var sqlOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"sqlmigrate", "blog", "0001_initial"}, &sqlOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("sqlmigrate error = %v", err)
	}
	if !strings.Contains(sqlOut.String(), `CREATE TABLE IF NOT EXISTS "blog_item"`) {
		t.Fatalf("sqlmigrate stdout = %q", sqlOut.String())
	}
}

func TestSQLMigrateUsesGeneratedOperationContent(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0001_initial",
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "EmptyMigration"}},
	})
	t.Chdir(dir)

	var stdout bytes.Buffer
	if err := NewRoot().Execute(context.Background(), []string{"sqlmigrate", "blog", "0001_initial"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("sqlmigrate error = %v", err)
	}
	if strings.Contains(stdout.String(), "CREATE TABLE") || !strings.Contains(stdout.String(), "No SQL operations") {
		t.Fatalf("sqlmigrate stdout = %q", stdout.String())
	}
}

func TestBuiltInAuthMigrationIsDiscoveredAndAppliedByDefault(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeMigrationEnv(t, dir, dbPath)
	t.Chdir(dir)

	root := NewRoot()
	var showOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"showmigrations", "--app", "auth"}, &showOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("showmigrations auth error = %v", err)
	}
	if !strings.Contains(showOut.String(), "[ ] auth.0001_initial") {
		t.Fatalf("showmigrations auth stdout = %q", showOut.String())
	}

	var sqlOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"sqlmigrate", "auth", "0001_initial"}, &sqlOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("sqlmigrate auth error = %v", err)
	}
	if !strings.Contains(sqlOut.String(), "CREATE TABLE auth_user") {
		t.Fatalf("sqlmigrate auth stdout = %q", sqlOut.String())
	}

	var migrateOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"migrate"}, &migrateOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("migrate error = %v", err)
	}
	if !strings.Contains(migrateOut.String(), "applied auth.0001_initial") {
		t.Fatalf("migrate stdout = %q", migrateOut.String())
	}
	for _, table := range []string{"gogo_content_type", "auth_permission", "auth_group", "auth_user"} {
		assertSQLiteTableExists(t, dbPath, table)
	}
	assertMigrationRecorded(t, dbPath, "auth", "0001_initial")
}

func TestMigrateFakeInitialRecordsExistingBuiltInAuthSchema(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeMigrationEnv(t, dir, dbPath)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	for _, statement := range sqlStatementsForMigration(authmigrations.Initial()) {
		if _, err := db.Exec(statement); err != nil {
			_ = db.Close()
			t.Fatalf("prepare existing auth schema: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	var stdout bytes.Buffer
	if err := NewRoot().Execute(context.Background(), []string{"migrate", "--app", "auth", "--fake-initial"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("migrate --fake-initial error = %v", err)
	}
	if !strings.Contains(stdout.String(), "applied auth.0001_initial") {
		t.Fatalf("migrate stdout = %q", stdout.String())
	}
	assertMigrationRecorded(t, dbPath, "auth", migrations.InitialMigrationName())
}

func TestMigrateAppliesGeneratedAppMigrationsAndShowMigrationsUsesRecorder(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeMigrationEnv(t, dir, dbPath)
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel: "blog",
		Name:     "0001_initial",
		Atomic:   true,
		Operations: []migrations.Operation{
			migrations.ManifestOperation{NameValue: "CreateModel:blog.Item"},
		},
	})
	t.Chdir(dir)

	root := NewRoot()
	var stdout bytes.Buffer
	if err := root.Execute(context.Background(), []string{"migrate"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("migrate error = %v", err)
	}
	if !strings.Contains(stdout.String(), "applied blog.0001_initial") {
		t.Fatalf("migrate stdout = %q", stdout.String())
	}
	assertSQLiteTableExists(t, dbPath, "blog_item")
	assertMigrationRecorded(t, dbPath, "blog", "0001_initial")

	var showOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"showmigrations"}, &showOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("showmigrations error = %v", err)
	}
	if !strings.Contains(showOut.String(), "[X] blog.0001_initial") {
		t.Fatalf("showmigrations stdout = %q", showOut.String())
	}
}

func TestMigrateUsesGeneratedOperationContent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeMigrationEnv(t, dir, dbPath)
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0001_initial",
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "EmptyMigration"}},
	})
	t.Chdir(dir)

	var stdout bytes.Buffer
	if err := NewRoot().Execute(context.Background(), []string{"migrate"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("migrate error = %v", err)
	}
	assertMigrationRecorded(t, dbPath, "blog", "0001_initial")
	assertSQLiteTableMissing(t, dbPath, "blog_item")
}

func TestMigratePlanListsPendingMigrations(t *testing.T) {
	dir := t.TempDir()
	writeMigrationEnv(t, dir, filepath.Join(dir, "db.sqlite3"))
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0001_initial",
		Atomic:     true,
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "CreateModel:blog.Item"}},
	})
	t.Chdir(dir)

	var stdout bytes.Buffer
	if err := NewRoot().Execute(context.Background(), []string{"migrate", "--plan"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("migrate --plan error = %v", err)
	}
	if !strings.Contains(stdout.String(), "apply blog.0001_initial") {
		t.Fatalf("migrate --plan stdout = %q", stdout.String())
	}
}

func TestMakeMigrationsCheckDryRunSkipsExistingMigration(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0001_initial",
		Atomic:     true,
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "CreateModel:blog.Item"}},
	})
	t.Chdir(dir)

	var stdout bytes.Buffer
	if err := NewRoot().Execute(context.Background(), []string{"makemigrations", "--app", "blog", "--name", "initial", "--check", "--dry-run"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("makemigrations --check --dry-run error = %v", err)
	}
	if !strings.Contains(stdout.String(), "no changes detected") || strings.Contains(stdout.String(), "would create") {
		t.Fatalf("makemigrations stdout = %q", stdout.String())
	}
}

func TestMakeMigrationsCheckFailsWhenMigrationWouldBeCreated(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "apps", "blog", "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	t.Chdir(dir)

	var stdout bytes.Buffer
	err := NewRoot().Execute(context.Background(), []string{"makemigrations", "--app", "blog", "--check", "--dry-run"}, &stdout, &bytes.Buffer{})
	if !errors.Is(err, ErrCommandFailed) {
		t.Fatalf("makemigrations error = %v, want ErrCommandFailed", err)
	}
	if !strings.Contains(stdout.String(), "would create blog.0001_initial") {
		t.Fatalf("makemigrations stdout = %q", stdout.String())
	}
}

func TestSquashMigrationsWritesReplacementMigrationFile(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0001_initial",
		Atomic:     true,
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "CreateModel:blog.Item"}},
	})
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0002_post",
		Atomic:     true,
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "AddField:blog.Item.slug"}},
	})
	t.Chdir(dir)

	root := NewRoot()
	var stdout bytes.Buffer
	if err := root.Execute(context.Background(), []string{"squashmigrations", "blog", "0001_initial", "0002_post", "--noinput"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("squashmigrations error = %v", err)
	}
	path := filepath.Join(migrationsDir, "0001_squashed_0002_post.go")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected squashed migration file %s: %v", path, err)
	}
	if !strings.Contains(string(contents), `Replaces: []gogomigrations.Dependency`) || !strings.Contains(string(contents), `Name: "0002_post"`) {
		t.Fatalf("squashed migration missing replacement metadata:\n%s", contents)
	}
	if !strings.Contains(stdout.String(), "created squashed migration blog.0001_squashed_0002_post replacing 2 migration(s)") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	writeTextFile(t, filepath.Join(dir, "go.mod"), "module generated-client\n\ngo 1.26.4\n\ntoolchain go1.26.4\n\nrequire github.com/cybersaksham/gogo v0.0.0\n")
	repoRoot := cliTestRepoRoot(t)
	runCLICommand(t, dir, "go", "mod", "edit", "-replace", "github.com/cybersaksham/gogo="+filepath.ToSlash(repoRoot))
	runCLICommand(t, dir, "go", "mod", "tidy")
	runCLICommand(t, dir, "go", "test", "./apps/blog/migrations")
}

func TestSquashedMigrationIsSatisfiedWhenReplacedMigrationsAreApplied(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeMigrationEnv(t, dir, dbPath)
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	writeGeneratedMigration(t, migrationsDir, migrations.Migration{
		AppLabel:   "blog",
		Name:       "0001_initial",
		Atomic:     true,
		Operations: []migrations.Operation{migrations.ManifestOperation{NameValue: "CreateModel:blog.Item"}},
	})
	t.Chdir(dir)

	root := NewRoot()
	if err := root.Execute(context.Background(), []string{"migrate"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("initial migrate error = %v", err)
	}
	if err := root.Execute(context.Background(), []string{"squashmigrations", "blog", "0001_initial", "0001_initial", "--noinput"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("squashmigrations error = %v", err)
	}

	var showOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"showmigrations"}, &showOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("showmigrations error = %v", err)
	}
	if !strings.Contains(showOut.String(), "[X] blog.0001_squashed_0001_initial (replaces applied: 0001_initial)") {
		t.Fatalf("showmigrations stdout = %q", showOut.String())
	}

	var planOut bytes.Buffer
	if err := root.Execute(context.Background(), []string{"migrate", "--plan"}, &planOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("migrate --plan error = %v", err)
	}
	if !strings.Contains(planOut.String(), "no migrations to apply") || strings.Contains(planOut.String(), "0001_squashed_0001_initial") {
		t.Fatalf("migrate --plan stdout = %q", planOut.String())
	}
}

func TestOptimizeMigrationReportsNoopWhenNoSafeRewriteExists(t *testing.T) {
	dir := t.TempDir()
	migrationsDir := filepath.Join(dir, "apps", "blog", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}
	writeTextFile(t, filepath.Join(migrationsDir, "0001_initial.go"), "package migrations\n")
	t.Chdir(dir)

	root := NewRoot()
	var stdout bytes.Buffer
	if err := root.Execute(context.Background(), []string{"optimizemigration", "blog", "0001_initial"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("optimizemigration error = %v", err)
	}
	if !strings.Contains(stdout.String(), "no optimizations needed for blog.0001_initial") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func writeMigrationEnv(t *testing.T, dir, dbPath string) {
	t.Helper()
	writeTextFile(t, filepath.Join(dir, ".env"), "GOGO_SECRET_KEY=migration-secret\nDATABASE_URL=sqlite://"+filepath.ToSlash(dbPath)+"\n")
}

func writeGeneratedMigration(t *testing.T, dir string, migration migrations.Migration) {
	t.Helper()
	if _, err := migrations.NewWriter(dir).Write(migration); err != nil {
		t.Fatalf("write migration %s: %v", migration.Identity(), err)
	}
}

func assertSQLiteTableExists(t *testing.T, dbPath, table string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	var name string
	if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name); err != nil {
		t.Fatalf("expected sqlite table %s: %v", table, err)
	}
}

func assertSQLiteTableMissing(t *testing.T, dbPath, table string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
	if err == nil {
		t.Fatalf("sqlite table %s exists", table)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("query sqlite table %s: %v", table, err)
	}
}

func assertMigrationRecorded(t *testing.T, dbPath, appLabel, name string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM gogo_migrations WHERE app = ? AND name = ?`, appLabel, name).Scan(&count); err != nil {
		t.Fatalf("query migration record: %v", err)
	}
	if count != 1 {
		t.Fatalf("migration record count = %d, want 1", count)
	}
}

func runCLICommand(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed in %s: %v\n%s", name, strings.Join(args, " "), dir, err, output)
	}
}

func cliTestRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
