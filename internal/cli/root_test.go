package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/internal/version"
)

func TestRootHelpListsPlannedCommandsInStableOrder(t *testing.T) {
	root := NewRoot()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := root.Execute(context.Background(), []string{"help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute(help) error = %v", err)
	}

	got := stdout.String()
	wantCommands := []string{
		"help",
		"version",
		"check",
		"runserver",
		"startproject",
		"startapp",
		"makemigrations",
		"migrate",
		"showmigrations",
		"sqlmigrate",
		"squashmigrations",
		"createsuperuser",
		"changepassword",
		"collectstatic",
		"shell",
		"dbshell",
		"test",
		"worker",
		"beat",
		"inspect",
		"queues",
		"dumpdata",
		"loaddata",
	}

	lastIndex := -1
	for _, command := range wantCommands {
		line := "  " + command + " "
		index := strings.Index(got, line)
		if index == -1 {
			t.Fatalf("help output missing command line containing %q:\n%s", line, got)
		}
		if index <= lastIndex {
			t.Fatalf("command %q was not listed after the previous command:\n%s", command, got)
		}
		lastIndex = index
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootVersionPrintsVersionInfo(t *testing.T) {
	restore := setVersionMetadata("9.8.7", "commit1", "2026-06-27T00:00:00Z")
	defer restore()

	root := NewRoot()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := root.Execute(context.Background(), []string{"version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute(version) error = %v", err)
	}

	want := "gogo 9.8.7 (commit commit1, built 2026-06-27T00:00:00Z)\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootUnavailableCommandReturnsPhaseError(t *testing.T) {
	root := NewRoot()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := root.Execute(context.Background(), []string{"runserver"}, &stdout, &stderr)
	if !errors.Is(err, ErrCommandUnavailable) {
		t.Fatalf("Execute(runserver) error = %v, want ErrCommandUnavailable", err)
	}

	if !strings.Contains(err.Error(), "03-http-routing-middleware-views") {
		t.Fatalf("Execute(runserver) error = %q, want phase name", err.Error())
	}
}

func TestRootUnknownCommandReturnsUnknownCommandError(t *testing.T) {
	root := NewRoot()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := root.Execute(context.Background(), []string{"missing"}, &stdout, &stderr)
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("Execute(missing) error = %v, want ErrUnknownCommand", err)
	}
}

func setVersionMetadata(versionValue, commit, buildDate string) func() {
	oldVersion := version.Version
	oldCommit := version.Commit
	oldBuildDate := version.BuildDate

	version.Version = versionValue
	version.Commit = commit
	version.BuildDate = buildDate

	return func() {
		version.Version = oldVersion
		version.Commit = oldCommit
		version.BuildDate = oldBuildDate
	}
}
