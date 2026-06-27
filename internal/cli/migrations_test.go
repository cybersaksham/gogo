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
