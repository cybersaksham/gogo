package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartprojectGeneratesExpectedFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "myproject")

	command := NewStartprojectCommand()
	if err := command.Run(context.Background(), []string{"myproject", target}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, path := range expectedProjectFiles("myproject") {
		fullPath := filepath.Join(target, path)
		if _, err := os.Stat(fullPath); err != nil {
			t.Fatalf("expected generated path %s: %v", path, err)
		}
	}
}

func TestStartprojectRejectsInvalidProjectName(t *testing.T) {
	command := NewStartprojectCommand()

	err := command.Run(context.Background(), []string{"bad-name", filepath.Join(t.TempDir(), "bad-name")})
	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("Run() error = %v, want ErrInvalidArguments", err)
	}
}

func TestStartprojectRefusesNonEmptyDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "myproject")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	writeTextFile(t, filepath.Join(target, "existing.txt"), "keep")

	command := NewStartprojectCommand()
	err := command.Run(context.Background(), []string{"myproject", target})
	if !errors.Is(err, ErrCommandFailed) {
		t.Fatalf("Run() error = %v, want ErrCommandFailed", err)
	}
}

func TestStartprojectForceAllowsNonEmptyDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "myproject")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	writeTextFile(t, filepath.Join(target, "existing.txt"), "keep")

	command := NewStartprojectCommand()
	if err := command.Run(context.Background(), []string{"--force", "myproject", target}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "existing.txt")); err != nil {
		t.Fatalf("existing file should be preserved by force generation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "manage.go")); err != nil {
		t.Fatalf("manage.go not generated: %v", err)
	}
}

func TestStartprojectEnvExampleContainsFrameworkKeys(t *testing.T) {
	target := filepath.Join(t.TempDir(), "myproject")

	command := NewStartprojectCommand()
	if err := command.Run(context.Background(), []string{"myproject", target}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	rootExample, err := os.ReadFile(filepath.Join("..", "..", ".env.example"))
	if err != nil {
		t.Fatalf("read root .env.example: %v", err)
	}
	generated, err := os.ReadFile(filepath.Join(target, ".env.example"))
	if err != nil {
		t.Fatalf("read generated .env.example: %v", err)
	}

	for _, line := range strings.Split(string(rootExample), "\n") {
		key, _, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || key == "" || strings.HasPrefix(key, "#") {
			continue
		}
		if !strings.Contains(string(generated), key+"=") {
			t.Fatalf("generated .env.example missing %s", key)
		}
	}
}

func expectedProjectFiles(projectName string) []string {
	return []string{
		"go.mod",
		"manage.go",
		".gitignore",
		".env.example",
		"Makefile",
		"README.md",
		filepath.Join(projectName, "app.go"),
		filepath.Join(projectName, "settings", "base.go"),
		filepath.Join(projectName, "settings", "local.go"),
		filepath.Join(projectName, "settings", "test.go"),
		filepath.Join(projectName, "settings", "production.go"),
		filepath.Join(projectName, "urls.go"),
		filepath.Join(projectName, "admin.go"),
		filepath.Join(projectName, "middleware.go"),
		filepath.Join(projectName, "queue.go"),
		filepath.Join("apps", ".keep"),
		filepath.Join("templates", "base.html"),
		filepath.Join("static", ".keep"),
		filepath.Join("media", ".keep"),
		filepath.Join("fixtures", ".keep"),
		filepath.Join("tests", "integration", ".keep"),
		filepath.Join("deploy", "docker", "Dockerfile"),
		filepath.Join("deploy", "docker", "docker-compose.yml"),
	}
}
