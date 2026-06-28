package templates

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestProjectFilesRenderExpectedStructure(t *testing.T) {
	files, err := ProjectFiles(ProjectData{ProjectName: "myproject", ModulePath: "myproject"})
	if err != nil {
		t.Fatalf("ProjectFiles() error = %v", err)
	}

	got := sortedKeys(files)
	want := []string{
		".env.example",
		".gitignore",
		"Makefile",
		"README.md",
		"apps/.keep",
		"deploy/docker/Dockerfile",
		"deploy/docker/docker-compose.yml",
		"fixtures/.keep",
		"go.mod",
		"manage.go",
		"media/.keep",
		filepath.Join("myproject", "admin.go"),
		filepath.Join("myproject", "app.go"),
		filepath.Join("myproject", "middleware.go"),
		filepath.Join("myproject", "queue.go"),
		filepath.Join("myproject", "settings", "base.go"),
		filepath.Join("myproject", "settings", "local.go"),
		filepath.Join("myproject", "settings", "production.go"),
		filepath.Join("myproject", "settings", "test.go"),
		filepath.Join("myproject", "urls.go"),
		"static/.keep",
		"templates/base.html",
		"tests/integration/.keep",
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("project files = %#v, want %#v", got, want)
	}
}

func TestProjectTemplatesRenderSafeConfigFiles(t *testing.T) {
	files, err := ProjectFiles(ProjectData{ProjectName: "myproject", ModulePath: "myproject"})
	if err != nil {
		t.Fatalf("ProjectFiles() error = %v", err)
	}

	gitignore := files[".gitignore"]
	for _, heading := range []string{
		"# Environment files",
		"# Go build outputs",
		"# Local databases",
		"# Uploaded media and collected static files",
		"# Coverage",
		"# Editor and OS files",
	} {
		if !strings.Contains(gitignore, heading) {
			t.Fatalf(".gitignore missing group heading %q", heading)
		}
	}

	envExample := files[".env.example"]
	for _, key := range []string{"GOGO_SECRET_KEY=", "DATABASE_URL=", "GOGO_HTTP_ADDR=:8000", "GOGO_ALLOWED_HOSTS=localhost,127.0.0.1"} {
		if !strings.Contains(envExample, key) {
			t.Fatalf(".env.example missing %s", key)
		}
	}
	if strings.Contains(envExample, "password") || strings.Contains(envExample, "secret-value") {
		t.Fatalf(".env.example must not contain committed secret placeholders")
	}
}

func TestProjectTemplatesRenderParseablePublicGoFiles(t *testing.T) {
	files, err := ProjectFiles(ProjectData{ProjectName: "myproject", ModulePath: "myproject"})
	if err != nil {
		t.Fatalf("ProjectFiles() error = %v", err)
	}

	for path, contents := range files {
		if filepath.Ext(path) != ".go" {
			continue
		}
		if strings.Contains(contents, "github.com/cybersaksham/gogo/internal") {
			t.Fatalf("%s imports internal framework package", path)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), path, contents, parser.AllErrors); err != nil {
			t.Fatalf("%s is not parseable Go: %v\n%s", path, err, contents)
		}
	}
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
