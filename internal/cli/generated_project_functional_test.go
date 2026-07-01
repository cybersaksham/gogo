//go:build integration

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/auth"
)

func TestGeneratedProjectFunctionalSurface(t *testing.T) {
	target := filepath.Join(t.TempDir(), "sampleproject")
	if err := NewStartprojectCommand().Run(context.Background(), []string{"sampleproject", target}); err != nil {
		t.Fatalf("startproject error = %v", err)
	}
	appTarget := filepath.Join(target, "apps", "blog")
	if err := NewStartappCommand().Run(context.Background(), []string{"blog", appTarget}); err != nil {
		t.Fatalf("startapp error = %v", err)
	}

	writeTextFile(t, filepath.Join(target, ".env"), `
GOGO_SECRET_KEY=functional-secret
DATABASE_URL=sqlite://./db.sqlite3
GOGO_HTTP_ADDR=127.0.0.1:0
`)

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	runGeneratedCommand(t, target, "go", "mod", "edit", "-replace", "github.com/cybersaksham/gogo="+filepath.ToSlash(repoRoot))
	writeTextFile(t, filepath.Join(target, "functional_test.go"), generatedFunctionalTestSource())
	runGeneratedCommand(t, target, "go", "mod", "tidy")

	withWorkingDirectory(t, filepath.Join(target, "apps"), func() {
		var stdout bytes.Buffer
		if err := NewRoot().Execute(context.Background(), []string{"makemigrations", "--app", "blog", "--name", "initial", "--check", "--dry-run"}, &stdout, &bytes.Buffer{}); err != nil {
			t.Fatalf("makemigrations check error = %v", err)
		}
		if !strings.Contains(stdout.String(), "no changes detected") {
			t.Fatalf("makemigrations stdout = %q", stdout.String())
		}
	})

	withWorkingDirectory(t, target, func() {
		var migrateStdout bytes.Buffer
		if err := NewRoot().Execute(context.Background(), []string{"migrate", "--database", "default"}, &migrateStdout, &bytes.Buffer{}); err != nil {
			t.Fatalf("migrate error = %v", err)
		}
		if !strings.Contains(migrateStdout.String(), "applied migrations on database default") {
			t.Fatalf("migrate stdout = %q", migrateStdout.String())
		}
	})

	store, _ := auth.NewMemoryUserStore()
	if err := NewCreateSuperuserCommand(store).Run(context.Background(), []string{
		"--username", "admin",
		"--email", "admin@example.com",
		"--password", "CorrectHorseBatteryStaple42",
		"--noinput",
	}); err != nil {
		t.Fatalf("createsuperuser error = %v", err)
	}
	user, ok, err := store.FindByUsername(context.Background(), "admin")
	if err != nil || !ok || !user.IsStaff || !user.IsSuperuser {
		t.Fatalf("created superuser = %#v, ok=%v, err=%v", user, ok, err)
	}

	withWorkingDirectory(t, target, func() {
		var captured RunserverConfig
		command := NewRunserverCommand(func(_ context.Context, config RunserverConfig) error {
			captured = config
			return nil
		})
		if err := command.Run(context.Background(), []string{"--addr", "127.0.0.1:0"}); err != nil {
			t.Fatalf("runserver error = %v", err)
		}
		if captured.Addr != "127.0.0.1:0" || captured.Settings.SecretKey == "" {
			t.Fatalf("captured runserver config = %#v", captured)
		}
	})

	runGeneratedCommand(t, target, "go", "test", "./...")
	assertNoInternalFrameworkImports(t, target)
}

func generatedFunctionalTestSource() string {
	return `package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	blog "sampleproject/apps/blog"
	project "sampleproject/sampleproject"

	"github.com/cybersaksham/gogo/admin"
	"github.com/cybersaksham/gogo/api"
	"github.com/cybersaksham/gogo/app"
	"github.com/cybersaksham/gogo/queue"
)

func TestGeneratedFunctionalSurface(t *testing.T) {
	router, err := project.NewRouter()
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("homepage status = %d", response.Code)
	}

	rawResponse := httptest.NewRecorder()
	router.ServeHTTP(rawResponse, httptest.NewRequest(http.MethodGet, "/raw/sample/", nil))
	if rawResponse.Code != http.StatusOK || rawResponse.Body.String() != "raw handler sample" {
		t.Fatalf("raw route response = %d body=%q", rawResponse.Code, rawResponse.Body.String())
	}

	site := project.NewAdminSite()
	if site.URLPrefix != "/admin" || site.ModelRegistry == nil {
		t.Fatalf("admin site = %#v", site)
	}

	apiRouter := api.NewRouter(api.WithAPIPrefix("api"))
	if err := blog.RegisterAPI(apiRouter); err != nil {
		t.Fatalf("RegisterAPI() error = %v", err)
	}
	apiResponse := httptest.NewRecorder()
	apiRouter.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/blog/items/", nil))
	if apiResponse.Code != http.StatusOK {
		t.Fatalf("api status = %d", apiResponse.Code)
	}

	if _, err := os.Stat("static/.keep"); err != nil {
		t.Fatalf("static path missing: %v", err)
	}

	appRegistry := app.NewRegistry()
	config := blog.NewConfig()
	if err := config.Ready(context.Background(), appRegistry); err != nil {
		t.Fatalf("app Ready() error = %v", err)
	}
	if len(appRegistry.Models()) == 0 || len(appRegistry.APIRoutes()) == 0 || len(appRegistry.Tasks()) == 0 || len(appRegistry.Migrations()) == 0 {
		t.Fatalf("app registry resources incomplete")
	}

	adminRegistry := admin.NewRegistry()
	if err := blog.RegisterAdmin(adminRegistry); err != nil {
		t.Fatalf("RegisterAdmin() error = %v", err)
	}
	if !adminRegistry.IsRegistered("blog.Item") {
		t.Fatalf("admin registry missing blog.Item")
	}

	queueApp := queue.NewApp(queue.AppOptions{})
	if err := blog.RegisterTasks(queueApp); err != nil {
		t.Fatalf("RegisterTasks() error = %v", err)
	}
	if _, ok := queueApp.Task("blog.example"); !ok {
		t.Fatalf("queue task not registered")
	}
}
`
}

func withWorkingDirectory(t *testing.T, dir string, fn func()) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()
	fn()
}
