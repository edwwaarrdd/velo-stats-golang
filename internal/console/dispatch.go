package console

import (
	"context"
	"flag"
	"fmt"

	"velostats/internal/app"
)

func dispatchTestTask() *Command {
	var message string

	return &Command{
		Name:        "tasks:dispatch-test",
		Description: "Dispatch a test job that logs a message from the worker.",
		Flags: func(flags *flag.FlagSet) {
			flags.StringVar(&message, "message", "Hello from tasks:dispatch-test", "Message the worker should write to the log")
		},
		Run: func(ctx context.Context, application *app.App) error {
			if err := application.Dispatcher.DispatchLogTestMessage(ctx, message); err != nil {
				return err
			}

			fmt.Println("Dispatched a test task to the queue.")

			return nil
		},
	}
}

func checkRideDistances() *Command {
	return &Command{
		Name:        "rides:check-distances",
		Description: "Queue a job per unchecked ride to calculate and cache the distance between its origin and destination stations.",
		Run: func(ctx context.Context, application *app.App) error {
			rideIDs, err := application.Rides.IDs(ctx, "distance_checked_at")
			if err != nil {
				return err
			}

			for _, rideID := range rideIDs {
				if err := application.Dispatcher.DispatchCheckRideDistance(ctx, rideID); err != nil {
					return err
				}
			}

			fmt.Printf("Dispatched %d ride distance check task(s).\n", len(rideIDs))

			return nil
		},
	}
}

func checkRideWeather() *Command {
	var force bool

	return &Command{
		Name:        "rides:check-weather",
		Description: "Queue a job per ride to fetch and cache the weather at its origin station and checkin time from Open-Meteo.",
		Flags: func(flags *flag.FlagSet) {
			flags.BoolVar(&force, "force", false, "Re-fetch weather for every ride, even if already checked")
		},
		Run: func(ctx context.Context, application *app.App) error {
			uncheckedColumn := "weather_checked_at"
			if force {
				uncheckedColumn = ""
			}

			rideIDs, err := application.Rides.IDs(ctx, uncheckedColumn)
			if err != nil {
				return err
			}

			for _, rideID := range rideIDs {
				if err := application.Dispatcher.DispatchCheckRideWeather(ctx, rideID, force); err != nil {
					return err
				}
			}

			fmt.Printf("Dispatched %d ride weather check task(s).\n", len(rideIDs))

			return nil
		},
	}
}
