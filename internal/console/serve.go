package console

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"velostats/internal/app"
	"velostats/internal/httpapi"
	"velostats/internal/queue"
)

func serve() *Command {
	return &Command{
		Name:        "serve",
		Description: "Serve the JSON API.",
		Run: func(ctx context.Context, application *app.App) error {
			server := &http.Server{
				Addr: ":" + application.Config.HTTPPort,
				Handler: httpapi.Router(
					application.Rides,
					application.RideSummary,
					application.RideCost,
					application.Stations,
					application.Config.CORSAllowedOrigins,
					application.Logger,
				),
				ReadHeaderTimeout: 5 * time.Second,
			}

			go func() {
				<-ctx.Done()

				shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				_ = server.Shutdown(shutdown)
			}()

			application.Logger.Info("serving the API", "port", application.Config.HTTPPort)

			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}
}

func work() *Command {
	var queues string

	return &Command{
		Name:        "work",
		Description: "Run a queue worker, draining the given queues one job at a time.",
		Flags: func(flags *flag.FlagSet) {
			flags.StringVar(&queues, "queue", queue.Default, "Comma-separated queues to consume")
		},
		Run: func(ctx context.Context, application *app.App) error {
			if err := application.Queue.Ping(ctx); err != nil {
				return fmt.Errorf("connect to redis: %w", err)
			}

			worker := queue.NewWorker(application.Queue, splitQueues(queues), application.DB, application.Logger)
			application.JobHandlers.Register(worker)

			return worker.Run(ctx)
		},
	}
}

func migrate() *Command {
	return &Command{
		Name:        "migrate",
		Description: "Create the database schema. Every command runs this first, so it is rarely needed on its own.",
		Run: func(_ context.Context, application *app.App) error {
			application.Logger.Info("the database schema is up to date")

			return nil
		},
	}
}
