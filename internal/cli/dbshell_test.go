package cli

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDBShellCommandResolvesSQLiteShell(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=dbshell-secret
DATABASE_URL=sqlite://./db.sqlite3
`)

	var got DBShellConfig
	command := NewDBShellCommand(func(_ context.Context, config DBShellConfig) error {
		got = config
		return nil
	})

	if err := command.Run(context.Background(), []string{"--command", "select 1"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got.Executable != "sqlite3" || len(got.Args) != 2 || got.Args[0] != "./db.sqlite3" || got.Args[1] != "select 1" {
		t.Fatalf("dbshell config = %#v", got)
	}
}

func TestDBShellDryRunRedactsPassword(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=dbshell-secret
DATABASE_URL=postgres://gogo:supersecret@localhost:5432/gogo?sslmode=disable
`)

	command := NewDBShellCommand(nil)
	runner := command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	})
	var stdout bytes.Buffer
	if err := runner.runWithIO(context.Background(), []string{"--dry-run"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "psql postgres://gogo@localhost:5432/gogo?sslmode=disable") || strings.Contains(output, "supersecret") {
		t.Fatalf("dry-run output = %q", output)
	}
}

func TestDBShellResolverMovesCredentialsToEnvironment(t *testing.T) {
	executable, args, env, err := resolveDBShellCommand("postgres://gogo:supersecret@localhost:5432/gogo?sslmode=disable", "", nil)
	if err != nil {
		t.Fatalf("resolve postgres error = %v", err)
	}
	if executable != "psql" || len(args) != 1 || strings.Contains(args[0], "supersecret") || !strings.Contains(args[0], "postgres://gogo@localhost:5432/gogo") {
		t.Fatalf("postgres command = %s %#v", executable, args)
	}
	if !reflect.DeepEqual(env, []string{"PGPASSWORD=supersecret"}) {
		t.Fatalf("postgres env = %#v", env)
	}

	executable, args, env, err = resolveDBShellCommand("mysql://gogo:secret@db.example.com:3306/app", "select 1", nil)
	if err != nil {
		t.Fatalf("resolve mysql error = %v", err)
	}
	if executable != "mysql" || strings.Contains(strings.Join(args, " "), "secret") || !reflect.DeepEqual(env, []string{"MYSQL_PWD=secret"}) {
		t.Fatalf("mysql command/env = %s %#v %#v", executable, args, env)
	}
}
