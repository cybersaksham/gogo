package management

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cybersaksham/gogo/internal/cli"
)

// Execute runs the Gogo management command registry.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return cli.NewRoot().Execute(ctx, args, stdout, stderr)
}

// Main runs management commands using os.Args and exits with a process status.
func Main() {
	if err := Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
