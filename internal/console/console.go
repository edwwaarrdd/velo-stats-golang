package console

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"velostats/internal/app"
)

type Command struct {
	Name        string
	Description string

	Flags func(flags *flag.FlagSet)

	Run func(ctx context.Context, application *app.App) error
}

func Commands() map[string]*Command {
	registered := []*Command{
		serve(),
		work(),
		migrate(),
		loadStations(),
		loadRides(),
		dispatchTestTask(),
		checkRideDistances(),
		checkRideWeather(),
	}

	byName := make(map[string]*Command, len(registered))
	for _, command := range registered {
		byName[command.Name] = command
	}

	return byName
}

func Usage(output *os.File) {
	commands := Commands()

	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintln(output, "Usage: velo <command> [options]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Commands:")

	writer := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	for _, name := range names {
		fmt.Fprintf(writer, "  %s\t%s\n", name, commands[name].Description)
	}
	writer.Flush()
}
