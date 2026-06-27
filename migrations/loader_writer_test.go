package migrations

import (
	"os"
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
	dir := t.TempDir()
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
	for _, want := range []string{"package migrations", `AppLabel: "blog"`, `"0002_add_post"`, `Dependencies:`} {
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
}
