// Command velo serves the Velo Stats API and runs its background work.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"velostats/internal/app"
	"velostats/internal/config"
	"velostats/internal/console"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		console.Usage(os.Stderr)

		return fmt.Errorf("no command given")
	}

	name := os.Args[1]

	if name == "-h" || name == "--help" || name == "help" {
		console.Usage(os.Stdout)

		return nil
	}

	command, known := console.Commands()[name]
	if !known {
		console.Usage(os.Stderr)

		return fmt.Errorf("unknown command %q", name)
	}

	flags := flag.NewFlagSet(name, flag.ExitOnError)
	if command.Flags != nil {
		command.Flags(flags)
	}

	if err := flags.Parse(os.Args[2:]); err != nil {
		return err
	}

	application, err := app.New(config.Load())
	if err != nil {
		return err
	}
	defer application.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return command.Run(ctx, application)
}
