package cli

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"
)

func TestTestCommandDefaultsToAllPackages(t *testing.T) {
	var got TestConfig
	command := NewTestCommand(func(_ context.Context, config TestConfig) error {
		got = config
		return nil
	})

	if err := command.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(got.Args, []string{"./..."}) {
		t.Fatalf("Args = %#v, want ./...", got.Args)
	}
}

func TestTestCommandPassesGoTestArgs(t *testing.T) {
	var got TestConfig
	command := NewTestCommand(func(_ context.Context, config TestConfig) error {
		got = config
		return nil
	})
	runner := command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	})

	var stdout bytes.Buffer
	if err := runner.runWithIO(context.Background(), []string{"./admin", "-run", "TestAdmin"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(got.Args, []string{"./admin", "-run", "TestAdmin"}) {
		t.Fatalf("Args = %#v", got.Args)
	}
	if got.Stdout != &stdout {
		t.Fatalf("Stdout was not passed to executor")
	}
}
