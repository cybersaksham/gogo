package management

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/auth"
	"github.com/cybersaksham/gogo/checks"
	"github.com/cybersaksham/gogo/internal/cli"
	"github.com/cybersaksham/gogo/queue"

	_ "modernc.org/sqlite"
)

func TestExecuteRunsBuiltInCommands(t *testing.T) {
	var stdout bytes.Buffer
	if err := Execute(context.Background(), []string{"help"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("Execute(help) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "runserver") || !strings.Contains(stdout.String(), "makemigrations") {
		t.Fatalf("help output missing management commands:\n%s", stdout.String())
	}
}

func TestExecuteProjectUsesProjectQueueApp(t *testing.T) {
	queueApp := queue.NewApp(queue.AppOptions{})
	_, err := queueApp.RegisterTask("blog.example", func(context.Context, ...any) (any, error) {
		return "ok", nil
	}, queue.TaskOptions{})
	if err != nil {
		t.Fatalf("RegisterTask() error = %v", err)
	}

	var stdout bytes.Buffer
	err = ExecuteProject(context.Background(), []string{"inspect", "--report"}, &stdout, &bytes.Buffer{}, Project{
		QueueApp: func() *queue.App {
			return queueApp
		},
	})
	if err != nil {
		t.Fatalf("ExecuteProject(inspect) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "registered=1") {
		t.Fatalf("inspect output missing registered task count:\n%s", stdout.String())
	}
}

func TestExecuteProjectRunsProjectCommand(t *testing.T) {
	var called bool
	var gotArgs []string
	command := projectCommand{
		name:    "blog.reindex",
		summary: "Reindex blog content",
		run: func(_ context.Context, args []string) error {
			called = true
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	var stdout bytes.Buffer
	err := ExecuteProject(context.Background(), []string{"blog.reindex", "--all"}, &stdout, &bytes.Buffer{}, Project{
		Commands: func() []Command {
			return []Command{command}
		},
	})
	if err != nil {
		t.Fatalf("ExecuteProject(custom command) error = %v", err)
	}
	if !called || strings.Join(gotArgs, ",") != "--all" {
		t.Fatalf("custom command called=%v args=%#v", called, gotArgs)
	}

	if err := ExecuteProject(context.Background(), []string{"check"}, &stdout, &bytes.Buffer{}, Project{
		Commands: func() []Command {
			return []Command{projectCommand{name: "check", summary: "Bad", run: func(context.Context, []string) error { return nil }}}
		},
	}); !errors.Is(err, cli.ErrDuplicateCommand) {
		t.Fatalf("ExecuteProject(duplicate command) error = %v, want ErrDuplicateCommand", err)
	}
}

func TestExecuteProjectRunsProjectChecks(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("GOGO_SECRET_KEY=management-secret\nDATABASE_URL=sqlite://:memory:\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	project := Project{
		Checks: func() []checks.Check {
			return []checks.Check{{
				ID:       "project.E001",
				Tags:     []string{"project"},
				Severity: checks.SeverityError,
				Message:  "PROJECT_REQUIRED is configured",
				Run: func(context.Context) checks.Result {
					if os.Getenv("PROJECT_REQUIRED") == "" {
						return checks.Result{ID: "project.E001", Tags: []string{"project"}, Severity: checks.SeverityError, Message: "PROJECT_REQUIRED is required"}
					}
					return checks.Result{ID: "project.I001", Tags: []string{"project"}, Severity: checks.SeverityInfo, Message: "project settings valid"}
				},
			}}
		},
	}

	var stdout bytes.Buffer
	err := ExecuteProject(context.Background(), []string{"check", "--tag", "project"}, &stdout, &bytes.Buffer{}, project)
	if !errors.Is(err, cli.ErrCommandFailed) {
		t.Fatalf("ExecuteProject(check) error = %v, want ErrCommandFailed", err)
	}
	if !strings.Contains(stdout.String(), "ERROR project PROJECT_REQUIRED is required") {
		t.Fatalf("check output missing project failure:\n%s", stdout.String())
	}

	t.Setenv("PROJECT_REQUIRED", "configured")
	stdout.Reset()
	if err := ExecuteProject(context.Background(), []string{"check", "--tag", "project"}, &stdout, &bytes.Buffer{}, project); err != nil {
		t.Fatalf("ExecuteProject(check configured) error = %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "INFO project project settings valid") {
		t.Fatalf("check output missing project pass:\n%s", stdout.String())
	}
}

func TestExecuteProjectCreateSuperuserPersistsToAuthUserTable(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	t.Chdir(dir)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("GOGO_SECRET_KEY=management-secret\nDATABASE_URL=sqlite://"+filepath.ToSlash(dbPath)+"\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	project := Project{ModelMetadata: auth.ModelMetadata}
	if err := ExecuteProject(context.Background(), []string{"migrate"}, &bytes.Buffer{}, &bytes.Buffer{}, project); err != nil {
		t.Fatalf("migrate error = %v", err)
	}
	var stdout bytes.Buffer
	if err := ExecuteProject(context.Background(), []string{
		"createsuperuser",
		"--username", "admin",
		"--email", "admin@example.com",
		"--password", "CorrectHorseBatteryStaple42",
		"--noinput",
	}, &stdout, &bytes.Buffer{}, project); err != nil {
		t.Fatalf("createsuperuser error = %v", err)
	}
	if !strings.Contains(stdout.String(), "created superuser admin on database default") {
		t.Fatalf("createsuperuser stdout = %q", stdout.String())
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	var email string
	var isStaff, isSuperuser bool
	if err := db.QueryRow(`SELECT email, is_staff, is_superuser FROM auth_user WHERE username = 'admin'`).Scan(&email, &isStaff, &isSuperuser); err != nil {
		t.Fatalf("query auth_user: %v", err)
	}
	if email != "admin@example.com" || !isStaff || !isSuperuser {
		t.Fatalf("auth_user row = email:%q staff:%v superuser:%v", email, isStaff, isSuperuser)
	}
}

type projectCommand struct {
	name    string
	summary string
	run     func(context.Context, []string) error
}

func (c projectCommand) Name() string {
	return c.name
}

func (c projectCommand) Summary() string {
	return c.summary
}

func (c projectCommand) Run(ctx context.Context, args []string) error {
	return c.run(ctx, args)
}
