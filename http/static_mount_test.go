package http

import (
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticMountServesFilesInDevelopment(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.css"), "body { color: black; }")

	handler, err := NewStaticMount(StaticMountConfig{
		Env:       "development",
		URLPrefix: "/static/",
		Root:      root,
	})
	if err != nil {
		t.Fatalf("NewStaticMount() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/static/app.css", nil))

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Body.String() != "body { color: black; }" {
		t.Fatalf("body = %q, want css content", recorder.Body.String())
	}
}

func TestStaticMountRefusesProductionByDefault(t *testing.T) {
	_, err := NewStaticMount(StaticMountConfig{
		Env:       "production",
		URLPrefix: "/media/",
		Root:      t.TempDir(),
	})
	if !errors.Is(err, ErrStaticMount) {
		t.Fatalf("NewStaticMount() error = %v, want ErrStaticMount", err)
	}
}

func TestStaticMountBlocksPathTraversal(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "static")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	writeFile(t, filepath.Join(base, "secret.txt"), "secret")

	handler, err := NewStaticMount(StaticMountConfig{
		Env:       "development",
		URLPrefix: "/static/",
		Root:      root,
	})
	if err != nil {
		t.Fatalf("NewStaticMount() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/static/%2e%2e/secret.txt", nil))

	if recorder.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}

func TestStaticMountReturnsNotFoundForMissingFiles(t *testing.T) {
	handler, err := NewStaticMount(StaticMountConfig{
		Env:       "development",
		URLPrefix: "/static/",
		Root:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("NewStaticMount() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/static/missing.css", nil))

	if recorder.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
