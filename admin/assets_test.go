package admin

import (
	"net/http"
	"net/http/httptest"
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
		"templates/password_change.html",
		"static/admin.css",
		"static/admin.js",
		"static/admin/css/base.css",
		"static/admin/css/changelists.css",
		"static/admin/css/forms.css",
		"static/admin/css/login.css",
		"static/admin/css/widgets.css",
		"static/admin/img/icon-addlink.svg",
		"static/admin/img/icon-changelink.svg",
		"static/admin/img/icon-deletelink.svg",
		"static/admin/img/search.svg",
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
	for _, want := range []string{
		`href="#content-start" class="skip-to-content-link"`,
		`<div class="main" id="main">`,
		`<main id="content-start" class="content" tabindex="-1">`,
		`<div id="content" class="colM">`,
		"Gogo administration",
		"blog",
		`id="recent-actions-module"`,
		`href="/admin/static/admin/css/base.css"`,
		`href="/admin/static/admin/css/dashboard.css"`,
		`href="/admin/static/admin/css/responsive.css"`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered index missing %q:\n%s", want, rendered)
		}
	}
}

func TestAdminStaticAssetsServeDjangoCSSAndImages(t *testing.T) {
	site := DefaultSite()
	router, err := site.URLs()
	if err != nil {
		t.Fatalf("URLs() error = %v", err)
	}

	tests := []struct {
		path        string
		contentType string
		want        string
	}{
		{"/admin/static/admin/css/base.css", "text/css; charset=utf-8", "--primary"},
		{"/admin/static/admin/css/changelists.css", "text/css; charset=utf-8", "#changelist"},
		{"/admin/static/admin/img/icon-addlink.svg", "image/svg+xml", "<svg"},
	}
	for _, test := range tests {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", test.path, response.Code, response.Body.String())
		}
		if got := response.Header().Get("Content-Type"); got != test.contentType {
			t.Fatalf("%s content type = %q", test.path, got)
		}
		if !strings.Contains(response.Body.String(), test.want) {
			t.Fatalf("%s body missing %q", test.path, test.want)
		}
	}
}

func TestAdminStaticAssetRejectsTraversal(t *testing.T) {
	site := DefaultSite()
	router, err := site.URLs()
	if err != nil {
		t.Fatalf("URLs() error = %v", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/admin/static/../templates/base.html", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("traversal status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestAdminTemplatesAllowOverrides(t *testing.T) {
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
