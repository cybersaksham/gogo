package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminEmbeddedAssetsExist(t *testing.T) {
	want := []string{
		"templates/base.html",
		"templates/index.html",
		"templates/login.html",
		"templates/change_list.html",
		"templates/change_form.html",
		"templates/delete_confirmation.html",
		"templates/history.html",
		"static/admin.css",
		"static/admin.js",
	}
	assets := AssetNames()
	for _, name := range want {
		if !containsString(assets, name) {
			t.Fatalf("asset %s missing from %#v", name, assets)
		}
		if body, ok := ReadAsset(name); !ok || len(body) == 0 {
			t.Fatalf("ReadAsset(%s) = %d, %v", name, len(body), ok)
		}
	}
}

func TestAdminTemplatesRenderBlocksAndAllowOverrides(t *testing.T) {
	rendered, err := RenderTemplate("index.html", map[string]any{"Header": "Gogo administration", "Apps": []IndexApp{{AppLabel: "blog"}}}, nil)
	if err != nil {
		t.Fatalf("RenderTemplate(index) error = %v", err)
	}
	if !strings.Contains(rendered, "<main") || !strings.Contains(rendered, "Gogo administration") || !strings.Contains(rendered, "blog") {
		t.Fatalf("rendered index = %s", rendered)
	}

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "admin", "templates"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "admin", "templates", "index.html"), []byte("override {{.Header}}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	override, err := RenderTemplate("index.html", map[string]any{"Header": "Custom"}, []string{dir})
	if err != nil {
		t.Fatalf("RenderTemplate(override) error = %v", err)
	}
	if override != "override Custom" {
		t.Fatalf("override = %q", override)
	}
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
