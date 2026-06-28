package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/cybersaksham/gogo/static"
)

// StaticCollector runs one collectstatic operation.
type StaticCollector func(context.Context, static.CollectOptions) (static.CollectResult, error)

// NewCollectstaticCommand creates the collectstatic command.
func NewCollectstaticCommand(collector StaticCollector) Command {
	if collector == nil {
		collector = static.Collect
	}
	return collectstaticCommand{collector: collector}
}

type collectstaticCommand struct {
	collector StaticCollector
}

func (c collectstaticCommand) Name() string {
	return "collectstatic"
}

func (c collectstaticCommand) Summary() string {
	return "Collect static files"
}

func (c collectstaticCommand) Run(ctx context.Context, args []string) error {
	return c.runWithIO(ctx, args, io.Discard, io.Discard)
}

func (c collectstaticCommand) runWithIO(ctx context.Context, args []string, stdout, _ io.Writer) error {
	options, err := parseCollectstaticFlags(args)
	if err != nil {
		return err
	}
	result, err := c.collector(ctx, options)
	if err != nil {
		return fmt.Errorf("%w: collect static files: %v", ErrCommandFailed, err)
	}
	if _, err := fmt.Fprintf(stdout, "collected %d static files\n", len(result.Copied)); err != nil {
		return fmt.Errorf("%w: write collectstatic output: %v", ErrCommandFailed, err)
	}
	return nil
}

func parseCollectstaticFlags(args []string) (static.CollectOptions, error) {
	flags := flag.NewFlagSet("collectstatic", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var projectDirs repeatedFlag
	var appDirs repeatedFlag
	options := static.CollectOptions{}
	flags.StringVar(&options.Destination, "dest", "", "collection destination")
	flags.BoolVar(&options.Manifest, "manifest", false, "write hashed manifest")
	flags.BoolVar(&options.Clear, "clear", false, "clear destination before collecting")
	flags.Var(&projectDirs, "project-dir", "project static directory")
	flags.Var(&appDirs, "app-dir", "app static directory")
	if err := flags.Parse(args); err != nil {
		return static.CollectOptions{}, fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}
	options.Finder.ProjectDirs = []string(projectDirs)
	options.Finder.AppDirs = []string(appDirs)
	return options, nil
}

type repeatedFlag []string

func (f *repeatedFlag) String() string {
	return fmt.Sprint([]string(*f))
}

func (f *repeatedFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}
