package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedProjectWithAppCompilesAsDownstreamModule(t *testing.T) {
	target := filepath.Join(t.TempDir(), "sampleproject")

	if err := NewStartprojectCommand().Run(context.Background(), []string{"sampleproject", target}); err != nil {
		t.Fatalf("startproject error = %v", err)
	}
	appTarget := filepath.Join(target, "apps", "blog")
	if err := NewStartappCommand().Run(context.Background(), []string{"blog", appTarget}); err != nil {
		t.Fatalf("startapp error = %v", err)
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	runGeneratedCommand(t, target, "go", "mod", "edit", "-replace", "github.com/cybersaksham/gogo="+filepath.ToSlash(repoRoot))
	writeTextFile(t, filepath.Join(target, "generated_runtime_test.go"), generatedRuntimeRouteTestSource())
	runGeneratedCommand(t, target, "go", "mod", "tidy")
	inspectOutput := runGeneratedCommandOutput(t, target, "go", "run", "manage.go", "inspect", "--report")
	if !strings.Contains(inspectOutput, "registered=1") {
		t.Fatalf("project-aware inspect output = %q, want registered task", inspectOutput)
	}
	runGeneratedCommand(t, target, "go", "test", "./...")
	assertNoInternalFrameworkImports(t, target)
}

func generatedRuntimeRouteTestSource() string {
	return `package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	project "sampleproject/sampleproject"
)

func TestGeneratedRouterMountsAppHTTPAPIAndAdminRoutes(t *testing.T) {
	router, err := project.NewRouter()
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	for _, test := range []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/blog/", http.StatusOK},
		{"/api/blog/items/", http.StatusOK},
		{"/admin/", http.StatusOK},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != test.want {
			t.Fatalf("%s status = %d body=%s, want %d", test.path, response.Code, response.Body.String(), test.want)
		}
	}
}
`
}

func assertNoInternalFrameworkImports(t *testing.T, root string) {
	t.Helper()
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(contents), "github.com/cybersaksham/gogo/internal") {
			t.Fatalf("generated file imports internal framework package: %s", path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk generated project: %v", err)
	}
}

func runGeneratedCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	output := runGeneratedCommandOutput(t, dir, name, args...)
	_ = output
}

func runGeneratedCommandOutput(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed in %s: %v\n%s", name, strings.Join(args, " "), dir, err, output)
	}
	return string(output)
}
