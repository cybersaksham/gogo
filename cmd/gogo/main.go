package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cybersaksham/gogo/internal/cli"
)

func main() {
	root := cli.NewRoot()
	if err := root.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, cli.ErrInvalidArguments) || errors.Is(err, cli.ErrUnknownCommand) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
