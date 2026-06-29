package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationCommandsRunWithFlags(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
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
