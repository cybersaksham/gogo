package migrations

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderLoadsMigrationManifests(t *testing.T) {
	dir := t.TempDir()
	if err := WriteManifest(dir, testMigration("blog", "0001_initial")); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	loader := NewLoader([]string{dir})
	migrations, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(migrations) != 1 || migrations[0].Identity() != "blog.0001_initial" {
		t.Fatalf("loaded migrations = %#v", migrations)
	}
}

func TestWriterWritesDeterministicGoMigration(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "blog", "migrations")
	migration := testMigration("blog", "0002_add_post")
	migration.Dependencies = []Dependency{{AppLabel: "blog", Name: "0001_initial"}}
	writer := NewWriter(dir)
	path, err := writer.Write(migration)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if filepath.Base(path) != "0002_add_post.go" {
		t.Fatalf("path = %q", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)
	for _, want := range []string{"package migrations", `gogomigrations.Migration`, `AppLabel: "blog"`, `"0002_add_post"`, `Dependencies:`} {
		if !strings.Contains(text, want) {
			t.Fatalf("written migration missing %q:\n%s", want, text)
		}
	}
	pathAgain, err := writer.Write(migration)
	if err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	again, _ := os.ReadFile(pathAgain)
	if string(again) != text {
		t.Fatalf("writer output was not deterministic")
	}
	second := testMigration("blog", "0003_add_comment")
	second.Dependencies = []Dependency{{AppLabel: "blog", Name: "0002_add_post"}}
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("Write(second) error = %v", err)
	}

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	writeTextFile(t, filepath.Join(root, "go.mod"), "module generated-migration\n\ngo 1.26.4\n\ntoolchain go1.26.4\n\nrequire github.com/cybersaksham/gogo v0.0.0\n")
	runTestCommand(t, root, "go", "mod", "edit", "-replace", "github.com/cybersaksham/gogo="+filepath.ToSlash(repoRoot))
	runTestCommand(t, root, "go", "mod", "tidy")
	runTestCommand(t, root, "go", "test", "./blog/migrations")
}

func writeTextFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runTestCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed in %s: %v\n%s", name, strings.Join(args, " "), dir, err, output)
	}
}
