package cli

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
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

func TestShellDefaultExecutorReturnsGuidance(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=shell-secret
DATABASE_URL=postgres://shell
`)

	command := NewShellCommand(nil)
	err := command.Run(context.Background(), []string{"--command", "print apps"})
	if !errors.Is(err, ErrCommandUnavailable) {
		t.Fatalf("Run() error = %v, want ErrCommandUnavailable", err)
	}
	if !strings.Contains(err.Error(), "interactive shell") {
		t.Fatalf("Run() error = %q, want shell guidance", err.Error())
	}
}
