package templates

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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
		filepath.Join(".agent", "rules", "gogo.md"),
		filepath.Join(".agent", "rules", "gogo", "forms-templates-static.md"),
		filepath.Join(".agent", "rules", "gogo", "http-admin-api-auth.md"),
		filepath.Join(".agent", "rules", "gogo", "models-orm-migrations.md"),
		filepath.Join(".agent", "rules", "gogo", "package-feature-index.md"),
		filepath.Join(".agent", "rules", "gogo", "project-structure.md"),
		filepath.Join(".agent", "rules", "gogo", "queue-workers.md"),
		filepath.Join(".agent", "rules", "gogo", "settings-security.md"),
		filepath.Join(".agent", "rules", "gogo", "testing-deployment.md"),
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

func TestProjectTemplatesRenderGogoAgentRules(t *testing.T) {
	files, err := ProjectFiles(ProjectData{ProjectName: "myproject", ModulePath: "myproject"})
	if err != nil {
		t.Fatalf("ProjectFiles() error = %v", err)
	}

	if _, ok := files["AGENTS.md"]; ok {
		t.Fatalf("generated project must not include root AGENTS.md")
	}
	if _, ok := files["CLAUDE.md"]; ok {
		t.Fatalf("generated project must not include root CLAUDE.md")
	}

	base := files[filepath.Join(".agent", "rules", "gogo.md")]
	for _, want := range []string{
		"Project type: Gogo client project.",
		".agent/rules/gogo/package-feature-index.md",
		".agent/rules/gogo/project-structure.md",
		".agent/rules/gogo/models-orm-migrations.md",
		".agent/rules/gogo/http-admin-api-auth.md",
		".agent/rules/gogo/forms-templates-static.md",
		".agent/rules/gogo/queue-workers.md",
		".agent/rules/gogo/settings-security.md",
		".agent/rules/gogo/testing-deployment.md",
	} {
		if !strings.Contains(base, want) {
			t.Fatalf("gogo.md missing %q:\n%s", want, base)
		}
	}

	index := files[filepath.Join(".agent", "rules", "gogo", "package-feature-index.md")]
	for _, want := range []string{
		"Public Package Map",
		"github.com/cybersaksham/gogo/contrib/sites",
		"gogo makemigrations",
		"gogo loaddata",
		"go doc github.com/cybersaksham/gogo/<package>",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("package-feature-index.md missing %q:\n%s", want, index)
		}
	}

	for path, contents := range files {
		if !strings.HasPrefix(path, filepath.Join(".agent", "rules", "gogo")) {
			continue
		}
		for _, want := range []string{
			"github.com/cybersaksham/gogo",
			"gogo",
		} {
			if !strings.Contains(contents, want) {
				t.Fatalf("%s missing %q:\n%s", path, want, contents)
			}
		}
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

	if !strings.Contains(files["manage.go"], "github.com/cybersaksham/gogo/management") {
		t.Fatalf("manage.go must use public management package:\n%s", files["manage.go"])
	}
	for _, want := range []string{
		`"myproject/myproject"`,
		`"myproject/myproject/settings"`,
		"management.MainProject",
		"Settings:   settings.Local",
		"AppConfigs: myproject.AppConfigs",
		"Router:     myproject.NewRouter",
		"QueueApp:   myproject.NewQueueApp",
	} {
		if !strings.Contains(files["manage.go"], want) {
			t.Fatalf("manage.go missing %q:\n%s", want, files["manage.go"])
		}
	}
}

func TestDeploymentTemplatesAreProductionSafe(t *testing.T) {
	files, err := ProjectFiles(ProjectData{ProjectName: "myproject", ModulePath: "myproject"})
	if err != nil {
		t.Fatalf("ProjectFiles() error = %v", err)
	}

	dockerfile := files["deploy/docker/Dockerfile"]
	for _, want := range []string{
		"FROM golang:1.26.4 AS build",
		"FROM gcr.io/distroless/static-debian12",
		"USER nonroot:nonroot",
		"CGO_ENABLED=0 go build",
	} {
		if !strings.Contains(dockerfile, want) {
			t.Fatalf("Dockerfile missing %q:\n%s", want, dockerfile)
		}
	}

	compose := files["deploy/docker/docker-compose.yml"]
	for _, want := range []string{
		"  app:",
		"  db:",
		"  redis:",
		"  rabbitmq:",
		"env_file:",
		"postgres-data:",
		"redis-data:",
		"rabbitmq-data:",
		"static-data:",
		"media-data:",
		"POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}",
		"profiles:",
	} {
		if !strings.Contains(compose, want) {
			t.Fatalf("compose template missing %q:\n%s", want, compose)
		}
	}

	for _, forbidden := range []string{
		"POSTGRES_PASSWORD: gogo",
		"guest:guest",
		"CorrectHorseBatteryStaple",
		"secret-value",
	} {
		if strings.Contains(compose, forbidden) || strings.Contains(files[".env.example"], forbidden) {
			t.Fatalf("deployment templates contain forbidden secret-like value %q", forbidden)
		}
	}
}

func TestAppFilesRenderExpectedStructure(t *testing.T) {
	files, err := AppFiles(AppData{AppName: "blog", AppLabel: "blog"})
	if err != nil {
		t.Fatalf("AppFiles() error = %v", err)
	}

	got := sortedKeys(files)
	want := []string{
		"admin.go",
		"api.go",
		"app.go",
		"forms.go",
		"migrations/.keep",
		"models.go",
		"permissions.go",
		"serializers.go",
		"services.go",
		"static/blog/.keep",
		"tasks.go",
		"templates/blog/.keep",
		"tests/blog_test.go",
		"urls.go",
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("app files = %#v, want %#v", got, want)
	}
}

func TestAppTemplatesRenderParseablePublicGoFiles(t *testing.T) {
	files, err := AppFiles(AppData{AppName: "blog", AppLabel: "blog"})
	if err != nil {
		t.Fatalf("AppFiles() error = %v", err)
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

func TestGeneratedAppCompilesAsDownstreamModule(t *testing.T) {
	files, err := AppFiles(AppData{AppName: "blog", AppLabel: "blog"})
	if err != nil {
		t.Fatalf("AppFiles() error = %v", err)
	}

	root := t.TempDir()
	writeFile(t, root, "go.mod", "module sample\n\ngo 1.26.4\n\ntoolchain go1.26.4\n\nrequire github.com/cybersaksham/gogo v0.0.0\n")
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	runCommand(t, root, "go", "mod", "edit", "-replace", "github.com/cybersaksham/gogo="+filepath.ToSlash(repoRoot))
	for path, contents := range files {
		writeFile(t, root, filepath.Join("apps", "blog", path), contents)
	}
	runCommand(t, root, "go", "mod", "tidy")
	runCommand(t, root, "go", "test", "./apps/blog/...")
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func writeFile(t *testing.T, root string, path string, contents string) {
	t.Helper()
	fullPath := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed in %s: %v\n%s", name, strings.Join(args, " "), dir, err, output)
	}
}
