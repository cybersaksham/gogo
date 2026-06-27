package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStartappGeneratesExpectedFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "blog")

	command := NewStartappCommand()
	if err := command.Run(context.Background(), []string{"blog", target}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, path := range expectedAppFiles("blog") {
		fullPath := filepath.Join(target, path)
		if _, err := os.Stat(fullPath); err != nil {
			t.Fatalf("expected generated path %s: %v", path, err)
		}
	}
}

func TestStartappRejectsInvalidAppName(t *testing.T) {
	command := NewStartappCommand()

	err := command.Run(context.Background(), []string{"bad-name", filepath.Join(t.TempDir(), "bad-name")})
	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("Run() error = %v, want ErrInvalidArguments", err)
	}
}

func TestStartappRefusesExistingDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "blog")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	writeTextFile(t, filepath.Join(target, "existing.txt"), "keep")

	command := NewStartappCommand()
	err := command.Run(context.Background(), []string{"blog", target})
	if !errors.Is(err, ErrCommandFailed) {
		t.Fatalf("Run() error = %v, want ErrCommandFailed", err)
	}
}

func TestStartappForceAllowsExistingDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "blog")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	writeTextFile(t, filepath.Join(target, "existing.txt"), "keep")

	command := NewStartappCommand()
	if err := command.Run(context.Background(), []string{"--force", "blog", target}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "existing.txt")); err != nil {
		t.Fatalf("existing file should be preserved by force generation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "app.go")); err != nil {
		t.Fatalf("app.go not generated: %v", err)
	}
}

func expectedAppFiles(appLabel string) []string {
	return []string{
		"app.go",
		"models.go",
		"admin.go",
		"urls.go",
		"api.go",
		"serializers.go",
		"forms.go",
		"services.go",
		"tasks.go",
		"permissions.go",
		filepath.Join("migrations", ".keep"),
		filepath.Join("templates", appLabel, ".keep"),
		filepath.Join("static", appLabel, ".keep"),
		filepath.Join("tests", ".keep"),
	}
}
