package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestStartappAutoInstallsIntoGeneratedProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "sampleproject")
	if err := NewStartprojectCommand().Run(context.Background(), []string{"sampleproject", root}); err != nil {
		t.Fatalf("startproject error = %v", err)
	}
	appTarget := filepath.Join(root, "apps", "blog")
	if err := NewStartappCommand().Run(context.Background(), []string{"blog", appTarget}); err != nil {
		t.Fatalf("startapp error = %v", err)
	}

	for path, wants := range map[string][]string{
		filepath.Join(root, "sampleproject", "settings", "base.go"): {
			`"blog",`,
		},
		filepath.Join(root, ".env.example"): {
			"GOGO_INSTALLED_APPS=gogo.contrib.sites,gogo.contrib.humanize,blog",
		},
		filepath.Join(root, "sampleproject", "urls.go"): {
			`"sampleproject/apps/blog"`,
			"blog.RegisterRoutes(router)",
			"NewAdminSite().URLs()",
		},
		filepath.Join(root, "sampleproject", "admin.go"): {
			`"sampleproject/apps/blog"`,
			"blog.RegisterAdmin(registry)",
		},
		filepath.Join(root, "sampleproject", "queue.go"): {
			`"sampleproject/apps/blog"`,
			"blog.RegisterTasks(app)",
		},
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, want := range wants {
			if !strings.Contains(string(contents), want) {
				t.Fatalf("%s missing %q:\n%s", path, want, contents)
			}
		}
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
		filepath.Join("tests", appLabel+"_test.go"),
	}
}
