package management

import (
	"bytes"
	"context"
	"strings"
	"testing"
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
