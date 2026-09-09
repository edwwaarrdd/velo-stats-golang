// Package app wires the application's pieces together.
package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"velostats/internal/config"
	"velostats/internal/database"
	"velostats/internal/jobs"
	"velostats/internal/queue"
	"velostats/internal/rides"
	"velostats/internal/routing"
	"velostats/internal/stations"
	"velostats/internal/weather"
)

// App holds every service the commands and the API need.
type App struct {
	Config config.Config
	Logger *slog.Logger
	DB     *sql.DB
	Queue  *queue.Client

	Rides          *rides.Repository
	Stations       *stations.Repository
	Routes         *routing.Repository
	WeatherRecords *weather.Repository

	RideSummary *rides.SummaryCalculator
	RideCost    *rides.CostCalculator

	RideSource         rides.DataSource
	StationInformation stations.InformationService
	CachedRoutes       *routing.CachedStationRouteService
	CachedWeather      *weather.CachedRideWeatherService

	Dispatcher  *jobs.Dispatcher
	JobHandlers *jobs.Handlers
}

// New builds the application and migrates its database.
func New(cfg config.Config) (*App, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db); err != nil {
		db.Close()

		return nil, err
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	rideRepository := rides.NewRepository(db)
	stationRepository := stations.NewRepository(db)
	routeRepository := routing.NewRepository(db)
	weatherRepository := weather.NewRepository(db)

	cachedRoutes := routing.NewCachedStationRouteService(
		routeRepository,
		routing.NewOSRMService(httpClient, cfg.OSRMBaseURL),
	)
	cachedWeather := weather.NewCachedRideWeatherService(
		weatherRepository,
		weather.NewOpenMeteoService(httpClient, cfg.OpenMeteoArchiveURL),
	)

	queueClient := queue.NewClient(redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}), cfg.QueuePrefix)

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Queue:  queueClient,

		Rides:          rideRepository,
		Stations:       stationRepository,
		Routes:         routeRepository,
		WeatherRecords: weatherRepository,

		RideSummary: rides.NewSummaryCalculator(db),
		RideCost:    rides.NewCostCalculator(rideRepository),

		RideSource:         rides.NewJSONFileService(cfg.RidesJSONPath),
		StationInformation: stations.NewVeloAntwerpService(httpClient, cfg.StationInformationURL),
		CachedRoutes:       cachedRoutes,
		CachedWeather:      cachedWeather,

		Dispatcher: jobs.NewDispatcher(queueClient),
		JobHandlers: jobs.NewHandlers(
			rideRepository,
			stationRepository,
			cachedRoutes,
			cachedWeather,
			logger,
		),
	}, nil
}

// Close releases the database and Redis connections.
func (a *App) Close() error {
	if err := a.Queue.Close(); err != nil {
		a.Logger.Error("failed to close the queue connection", "error", err)
	}

	return a.DB.Close()
}

// Context returns a context that is cancelled when the process is interrupted.
func Context(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}
