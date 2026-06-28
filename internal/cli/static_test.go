package cli

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/cybersaksham/gogo/static"
)

func TestCollectstaticCommandRunsConfiguredCollector(t *testing.T) {
	var captured static.CollectOptions
	command := NewCollectstaticCommand(func(_ context.Context, options static.CollectOptions) (static.CollectResult, error) {
		captured = options
		return static.CollectResult{Copied: []static.CollectedFile{{SourcePath: "css/app.css"}, {SourcePath: "js/app.js"}}}, nil
	})

	var stdout bytes.Buffer
	err := command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	}).runWithIO(context.Background(), []string{
		"--dest", "public/static",
		"--project-dir", "assets",
		"--app-dir", "blog/static",
		"--manifest",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runWithIO() error = %v", err)
	}
	if captured.Destination != "public/static" || !captured.Manifest {
		t.Fatalf("captured options = %#v", captured)
	}
	if !reflect.DeepEqual(captured.Finder.ProjectDirs, []string{"assets"}) || !reflect.DeepEqual(captured.Finder.AppDirs, []string{"blog/static"}) {
		t.Fatalf("captured finder = %#v", captured.Finder)
	}
	if stdout.String() != "collected 2 static files\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
