package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/cybersaksham/gogo/checks"
	"github.com/cybersaksham/gogo/conf"
	"github.com/cybersaksham/gogo/internal/release"
)

// NewCheckCommand creates the built-in system check command.
func NewCheckCommand() Command {
	return checkCommand{}
}

type checkCommand struct{}

func (c checkCommand) Name() string {
	return "check"
}

func (c checkCommand) Summary() string {
	return "Run system checks"
}

func (c checkCommand) Run(ctx context.Context, args []string) error {
	settings, err := conf.LoadFromEnv()
	if err != nil {
		return err
	}
	if err := settings.Validate(); err != nil {
		return err
	}
	options, failLevel, deploy, err := parseCheckFlags(args)
	if err != nil {
		return err
	}
	results := checks.DefaultRegistry().Run(ctx, options)
	if deploy {
		results = append(results, release.RunDeployChecks(release.BuildDeployConfig(ctx, settings))...)
	}
	if checks.HasFailures(results, failLevel) {
		return ErrCommandFailed
	}
	return nil
}

func (c checkCommand) runWithIO(ctx context.Context, args []string, stdout, _ io.Writer) error {
	options, failLevel, deploy, err := parseCheckFlags(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCommandFailed, err)
	}
	settings, err := conf.LoadFromEnv()
	if err != nil {
		return err
	}

	if err := settings.Validate(); err != nil {
		fmt.Fprintf(stdout, "ERROR config %v\n", err)
		return err
	}

	fmt.Fprintln(stdout, "OK config settings valid")
	results := checks.DefaultRegistry().Run(ctx, options)
	if deploy {
		results = append(results, release.RunDeployChecks(release.BuildDeployConfig(ctx, settings))...)
	}
	for _, result := range results {
		tag := "general"
		if len(result.Tags) > 0 {
			tag = result.Tags[0]
		}
		fmt.Fprintf(stdout, "%s %s %s\n", result.Severity.String(), tag, result.Message)
		if result.Hint != "" {
			fmt.Fprintf(stdout, "HINT %s %s\n", result.ID, result.Hint)
		}
	}
	if checks.HasFailures(results, failLevel) {
		return ErrCommandFailed
	}

	return nil
}

func parseCheckFlags(args []string) (checks.Options, checks.Severity, bool, error) {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var tags string
	var failLevel string
	var deploy bool
	flags.StringVar(&tags, "tag", "", "comma-separated check tags")
	flags.StringVar(&failLevel, "fail-level", "ERROR", "minimum severity that fails the command")
	flags.BoolVar(&deploy, "deploy", false, "run production deployment checks")
	if err := flags.Parse(args); err != nil {
		return checks.Options{}, checks.SeverityError, false, err
	}
	level := parseSeverity(failLevel)
	return checks.Options{Tags: splitCheckTags(tags)}, level, deploy, nil
}

func splitCheckTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(tags, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseSeverity(value string) checks.Severity {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "DEBUG":
		return checks.SeverityDebug
	case "INFO":
		return checks.SeverityInfo
	case "WARN", "WARNING":
		return checks.SeverityWarning
	case "CRITICAL":
		return checks.SeverityCritical
	default:
		return checks.SeverityError
	}
}
