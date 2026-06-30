package management

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/auth"
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
