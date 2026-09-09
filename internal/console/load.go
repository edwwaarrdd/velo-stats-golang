package console

import (
	"context"
	"flag"
	"fmt"

	"velostats/internal/app"
	"velostats/internal/rides"
)

func loadStations() *Command {
	return &Command{
		Name:        "stations:load",
		Description: "Fetch Velo Antwerp station information and save it to the database.",
		Run: func(ctx context.Context, application *app.App) error {
			fetched, err := application.StationInformation.FetchStations(ctx)
			if err != nil {
				return err
			}

			created, updated := 0, 0

			for _, station := range fetched {
				wasCreated, err := application.Stations.Save(ctx, station)
				if err != nil {
					return err
				}

				if wasCreated {
					created++
				} else {
					updated++
				}
			}

			fmt.Printf("Loaded %d stations (%d created, %d updated).\n", len(fetched), created, updated)

			return nil
		},
	}
}

func loadRides() *Command {
	var path string

	return &Command{
		Name:        "rides:load",
		Description: "Load customer ride history from a JSON export into the database.",
		Flags: func(flags *flag.FlagSet) {
			flags.StringVar(&path, "path", "", "Path to the rides JSON export (defaults to the configured export)")
		},
		Run: func(ctx context.Context, application *app.App) error {
			source := application.RideSource
			if path != "" {
				source = rides.NewJSONFileService(path)
			}

			fetched, err := source.FetchRides()
			if err != nil {
				return err
			}

			created, updated := 0, 0

			for _, ride := range fetched {
				wasCreated, err := application.Rides.Save(ctx, ride)
				if err != nil {
					return err
				}

				if wasCreated {
					created++
				} else {
					updated++
				}
			}

			fmt.Printf("Loaded %d rides (%d created, %d updated).\n", len(fetched), created, updated)

			return nil
		},
	}
}
