package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunDelegatesProjectAwareCommandToManageGo(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.mod"), "module sampleproject\n\ngo 1.26.4\n")
	writeTestFile(t, filepath.Join(root, "manage.go"), "package main\n")
	appDir := filepath.Join(root, "apps", "notes")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Chdir(appDir)

	var executed []string
	err := run(context.Background(), []string{"dumpdata", "notes.Item"}, &bytes.Buffer{}, &bytes.Buffer{}, runOptions{
		commandRunner: func(_ context.Context, dir, name string, args []string, _, _ io.Writer) error {
			executed = append([]string{name}, args...)
			if dir != root {
				t.Fatalf("delegate dir = %q, want %q", dir, root)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	want := []string{"go", "run", "manage.go", "dumpdata", "notes.Item"}
	if !reflect.DeepEqual(executed, want) {
		t.Fatalf("executed = %#v, want %#v", executed, want)
	}
}

func TestRunDoesNotDelegateGlobalOnlyCommands(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "go.mod"), "module sampleproject\n\ngo 1.26.4\n")
	writeTestFile(t, filepath.Join(root, "manage.go"), "package main\n")
	t.Chdir(root)

	var called bool
	var stdout bytes.Buffer
	err := run(context.Background(), []string{"--version"}, &stdout, &bytes.Buffer{}, runOptions{
		commandRunner: func(context.Context, string, string, []string, io.Writer, io.Writer) error {
			called = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if called {
		t.Fatal("global-only version command should not delegate to manage.go")
	}
	if stdout.String() == "" {
		t.Fatal("version output is empty")
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
