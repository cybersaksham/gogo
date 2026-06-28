package templates

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadTemplatesPrecedenceProjectOverridesAppOverridesFramework(t *testing.T) {
	projectDir := t.TempDir()
	appDir := t.TempDir()

	writeTemplateFile(t, projectDir, "shared.html", "project")
	writeTemplateFile(t, appDir, "shared.html", "app")
	writeTemplateFile(t, appDir, "blog/post.html", "app-post")

	loaded, err := LoadTemplates(LoaderConfig{
		ProjectDirs:        []string{projectDir},
		AppDirs:            []string{appDir},
		FrameworkTemplates: map[string]string{"shared.html": "framework", "admin/base.html": "framework-admin"},
	})
	if err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}
	if loaded["shared.html"] != "project" {
		t.Fatalf("shared.html = %q, want project", loaded["shared.html"])
	}
	if loaded["blog/post.html"] != "app-post" {
		t.Fatalf("blog/post.html = %q, want app-post", loaded["blog/post.html"])
	}
	if loaded["admin/base.html"] != "framework-admin" {
		t.Fatalf("admin/base.html = %q, want framework-admin", loaded["admin/base.html"])
	}

	engine := NewEngine(WithTemplates(loaded))
	rendered, err := engine.Render("shared.html", nil)
	if err != nil {
		t.Fatalf("Render(shared.html) error = %v", err)
	}
	if rendered != "project" {
		t.Fatalf("rendered shared.html = %q", rendered)
	}
}

func TestTemplateHelpers(t *testing.T) {
	engine := NewEngine(
		WithTemplates(map[string]string{
			"page": `{{url "post" 7}}|{{static "css/app.css"}}|{{media "avatar.png"}}|{{date .when "2006-01-02"}}|{{default .empty "fallback"}}|{{length .items}}|{{join .items ","}}|{{pluralize 2 "item" "items"}}|{{linebreaks .text}}|{{safe_escape .unsafe}}`,
		}),
		WithTemplateHelpers(HelperConfig{
			StaticURL: "/static/",
			MediaURL:  "/media/",
			URLResolver: func(name string, args ...any) (string, error) {
				if name == "post" && len(args) == 1 && args[0] == 7 {
					return "/posts/7/", nil
				}
				return "", ErrTemplateNotFound
			},
		}),
	)

	rendered, err := engine.Render("page", Context{
		"when":   time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC),
		"items":  []string{"go", "api"},
		"text":   "line <one>\nline two",
		"unsafe": `<x>`,
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := `/posts/7/|/static/css/app.css|/media/avatar.png|2026-06-28|fallback|2|go,api|items|<p>line &lt;one&gt;<br>line two</p>|&lt;x&gt;`
	if rendered != want {
		t.Fatalf("Render() = %q, want %q", rendered, want)
	}
}

func writeTemplateFile(t *testing.T, root, name, source string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
