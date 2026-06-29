package management

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/queue"
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
