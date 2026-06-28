package cli

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"
)

func TestShellCommandExecutesNonInteractiveCommand(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=shell-secret
DATABASE_URL=postgres://shell
`)

	var got ShellConfig
	command := NewShellCommand(func(_ context.Context, config ShellConfig) error {
		got = config
		return nil
	})

	err := command.Run(context.Background(), []string{"--command", "print apps"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if got.Command != "print apps" {
		t.Fatalf("Command = %q, want print apps", got.Command)
	}
	if got.Settings.SecretKey != "shell-secret" {
		t.Fatalf("Settings.SecretKey = %q, want shell-secret", got.Settings.SecretKey)
	}
	if got.Registry == nil {
		t.Fatalf("Registry = nil, want empty registry")
	}
}

func TestShellDefaultExecutorRunsCommand(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=shell-secret
DATABASE_URL=postgres://shell
`)

	command := NewShellCommand(nil)
	runner, ok := command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	})
	if !ok {
		t.Fatal("shell command does not expose runWithIO")
	}
	var stdout bytes.Buffer
	if err := runner.runWithIO(context.Background(), []string{"--command", "printf shell-ok"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if stdout.String() != "shell-ok" {
		t.Fatalf("stdout = %q, want shell-ok", stdout.String())
	}
}
