package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/conf"
)

func TestCheckCommandPassesValidConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=check-secret
DATABASE_URL=postgres://check
`)

	root := NewRoot()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := root.Execute(context.Background(), []string{"check"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute(check) error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"OK config settings valid",
		"WARN apps app registry checks unavailable until phase 02-app-project-lifecycle",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("check output = %q, want it to contain %q", output, want)
		}
	}
}

func TestCheckCommandFailsInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	root := NewRoot()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := root.Execute(context.Background(), []string{"check"}, &stdout, &stderr)
	if !errors.Is(err, conf.ErrInvalidSettings) {
		t.Fatalf("Execute(check) error = %v, want ErrInvalidSettings", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"ERROR config",
		"GOGO_SECRET_KEY",
		"DATABASE_URL",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("check output = %q, want it to contain %q", output, want)
		}
	}
}

func TestCheckCommandSupportsTagFiltering(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=check-secret
DATABASE_URL=postgres://check
`)
	root := NewRoot()
	var stdout bytes.Buffer
	if err := root.Execute(context.Background(), []string{"check", "--tag", "queue"}, &stdout, io.Discard); err != nil {
		t.Fatalf("Execute(check --tag queue) error = %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "INFO queue queue checks registered") || strings.Contains(output, "WARN apps") {
		t.Fatalf("filtered check output = %q", output)
	}
}

func writeTextFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
