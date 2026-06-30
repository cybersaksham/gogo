package static

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindStaticFilesPrecedenceAndDuplicates(t *testing.T) {
	projectDir := t.TempDir()
	appDir := t.TempDir()
	writeStaticFile(t, projectDir, "css/shared.css", "project")
	writeStaticFile(t, appDir, "css/shared.css", "app")
	writeStaticFile(t, appDir, "js/app.js", "app-js")

	files, duplicates, err := Find(FinderConfig{
		ProjectDirs:    []string{projectDir},
		AppDirs:        []string{appDir},
		FrameworkFiles: map[string][]byte{"css/shared.css": []byte("framework"), "admin/base.css": []byte("framework-admin")},
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if got := sourceContent(files, "css/shared.css"); got != "project" {
		t.Fatalf("css/shared.css winner = %q, want project", got)
	}
	if got := sourceContent(files, "js/app.js"); got != "app-js" {
		t.Fatalf("js/app.js winner = %q, want app-js", got)
	}
	if got := sourceContent(files, "admin/base.css"); got != "framework-admin" {
		t.Fatalf("admin/base.css winner = %q, want framework-admin", got)
	}
	if len(duplicates) != 1 || duplicates[0].Path != "css/shared.css" || duplicates[0].Winner.Origin != "project" || len(duplicates[0].Losers) != 2 {
		t.Fatalf("duplicates = %#v", duplicates)
	}
}

func TestManifestHashing(t *testing.T) {
	file := FoundFile{Path: "css/app.css", Content: []byte("body{}")}
	hashed := HashedName(file.Path, file.Content)
	if !strings.HasPrefix(hashed, "css/app.") || !strings.HasSuffix(hashed, ".css") || hashed == file.Path {
		t.Fatalf("HashedName() = %q", hashed)
	}
	manifest := BuildManifest([]FoundFile{file})
	if manifest[file.Path] != hashed {
		t.Fatalf("manifest = %#v, want %q", manifest, hashed)
	}
}

func TestCollectStaticWritesHashedFilesAndManifest(t *testing.T) {
	projectDir := t.TempDir()
	destination := t.TempDir()
	writeStaticFile(t, projectDir, "css/app.css", "body{}")

	result, err := Collect(context.Background(), CollectOptions{
		Finder:      FinderConfig{ProjectDirs: []string{projectDir}},
		Destination: destination,
		Manifest:    true,
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(result.Copied) != 1 {
		t.Fatalf("Copied length = %d, want 1", len(result.Copied))
	}
	copied := result.Copied[0]
	if copied.SourcePath != "css/app.css" || copied.OutputPath == "css/app.css" {
		t.Fatalf("copied = %#v", copied)
	}
	content, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(copied.OutputPath)))
	if err != nil || string(content) != "body{}" {
		t.Fatalf("collected content = %q, %v", content, err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(destination, "staticfiles.json"))
	if err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest JSON error = %v", err)
	}
	if manifest["css/app.css"] != copied.OutputPath {
		t.Fatalf("manifest = %#v, copied=%#v", manifest, copied)
	}
}

func TestCollectStaticDryRunDiscoversFilesWithoutWriting(t *testing.T) {
	projectDir := t.TempDir()
	destination := filepath.Join(t.TempDir(), "staticfiles")
	writeStaticFile(t, projectDir, "css/app.css", "body{}")

	result, err := Collect(context.Background(), CollectOptions{
		Finder:      FinderConfig{ProjectDirs: []string{projectDir}},
		Destination: destination,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("Collect() dry-run error = %v", err)
	}
	if len(result.Copied) != 1 || result.Copied[0].SourcePath != "css/app.css" {
		t.Fatalf("Copied = %#v, want discovered css/app.css", result.Copied)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("dry-run destination stat error = %v, want not exist", err)
	}
}

func writeStaticFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func sourceContent(files []FoundFile, path string) string {
	for _, file := range files {
		if file.Path == path {
			return string(file.Content)
		}
	}
	return ""
}
