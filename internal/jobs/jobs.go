// Package jobs holds the background work the queues carry, and the handlers
// that run it.
package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"velostats/internal/queue"
	"velostats/internal/rides"
	"velostats/internal/routing"
	"velostats/internal/stations"
	"velostats/internal/support"
	"velostats/internal/weather"
)

// The job types carried on the queues.
const (
	TypeLogTestMessage    = "log_test_message"
	TypeCheckRideDistance = "check_ride_distance"
	TypeCheckRideWeather  = "check_ride_weather"
)

// LogTestMessagePayload is the message a test job asks the worker to log.
type LogTestMessagePayload struct {
	Message string `json:"message"`
}

// CheckRideDistancePayload names the ride whose distance should be resolved.
type CheckRideDistancePayload struct {
	RideID int64 `json:"ride_id"`
}

// CheckRideWeatherPayload names the ride whose weather should be fetched.
type CheckRideWeatherPayload struct {
	RideID int64 `json:"ride_id"`
	Force  bool  `json:"force"`
}

// Dispatcher pushes jobs onto their queues.
type Dispatcher struct {
	queue *queue.Client
}

// NewDispatcher builds a dispatcher over the given queue client.
func NewDispatcher(client *queue.Client) *Dispatcher {
	return &Dispatcher{queue: client}
}

// DispatchLogTestMessage queues a test job that logs a message from the worker.
func (d *Dispatcher) DispatchLogTestMessage(ctx context.Context, message string) error {
	return d.queue.Push(ctx, queue.Default, TypeLogTestMessage, LogTestMessagePayload{Message: message})
}

// DispatchCheckRideDistance queues a distance check for a ride.
func (d *Dispatcher) DispatchCheckRideDistance(ctx context.Context, rideID int64) error {
	return d.queue.Push(ctx, queue.RideDistanceChecks, TypeCheckRideDistance, CheckRideDistancePayload{RideID: rideID})
}

// DispatchCheckRideWeather queues a weather check for a ride.
func (d *Dispatcher) DispatchCheckRideWeather(ctx context.Context, rideID int64, force bool) error {
	return d.queue.Push(ctx, queue.RideWeatherChecks, TypeCheckRideWeather, CheckRideWeatherPayload{RideID: rideID, Force: force})
}

// Handlers runs the jobs the queues carry.
type Handlers struct {
	rides    *rides.Repository
	stations *stations.Repository
	routes   *routing.CachedStationRouteService
	weather  *weather.CachedRideWeatherService
	logger   *slog.Logger
}

// NewHandlers builds the job handlers.
func NewHandlers(
	rideRepository *rides.Repository,
	stationRepository *stations.Repository,
	routes *routing.CachedStationRouteService,
	weatherService *weather.CachedRideWeatherService,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		rides:    rideRepository,
		stations: stationRepository,
		routes:   routes,
		weather:  weatherService,
		logger:   logger,
	}
}

// Register attaches every handler to a worker.
func (h *Handlers) Register(worker *queue.Worker) {
	worker.Register(TypeLogTestMessage, h.LogTestMessage)
	worker.Register(TypeCheckRideDistance, h.CheckRideDistance)
	worker.Register(TypeCheckRideWeather, h.CheckRideWeather)
}

// LogTestMessage logs a message from the worker, so the queue setup can be
// verified.
func (h *Handlers) LogTestMessage(_ context.Context, raw json.RawMessage) error {
	var payload LogTestMessagePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("decode log test message payload: %w", err)
	}

	h.logger.Info("Test task received: " + payload.Message)

	return nil
}

// CheckRideDistance calculates and caches the cycling distance between a ride's
// origin and destination stations.
func (h *Handlers) CheckRideDistance(ctx context.Context, raw json.RawMessage) error {
	var payload CheckRideDistancePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("decode ride distance payload: %w", err)
	}

	ride, err := h.rides.Find(ctx, payload.RideID)
	if err != nil {
		return err
	}

	if ride == nil {
		return fmt.Errorf("ride %d not found", payload.RideID)
	}

	if ride.DistanceCheckedAt != nil {
		return nil
	}

	origin, destination, err := h.stationPair(ctx, ride.OriginStationCode, ride.DestinationStationCode)
	if err != nil {
		return err
	}

	if origin == nil || destination == nil {
		h.logger.Error(fmt.Sprintf(
			"Cannot check distance for ride %d: unknown station code(s) %s / %s",
			payload.RideID, ride.OriginStationCode, ride.DestinationStationCode,
		))

		return nil
	}

	if _, err := h.routes.GetRoute(ctx, *origin, *destination, routing.ModeBike); err != nil {
		return err
	}

	return h.rides.MarkChecked(ctx, ride.RideID, "distance_checked_at", time.Now())
}

// CheckRideWeather fetches and caches the weather at a ride's origin station and
// checkin time.
func (h *Handlers) CheckRideWeather(ctx context.Context, raw json.RawMessage) error {
	var payload CheckRideWeatherPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("decode ride weather payload: %w", err)
	}

	ride, err := h.rides.Find(ctx, payload.RideID)
	if err != nil {
		return err
	}

	if ride == nil {
		return fmt.Errorf("ride %d not found", payload.RideID)
	}

	if ride.WeatherCheckedAt != nil && !payload.Force {
		return nil
	}

	origin, err := h.stations.Find(ctx, ride.OriginStationCode)
	if err != nil {
		return err
	}

	if origin == nil {
		h.logger.Error(fmt.Sprintf(
			"Cannot check weather for ride %d: unknown origin station code %s",
			payload.RideID, ride.OriginStationCode,
		))

		return nil
	}

	if _, err := h.weather.GetWeather(
		ctx,
		ride.RideID,
		ride.CheckinTime,
		support.Coordinate{Lat: origin.Lat, Lon: origin.Lon},
		payload.Force,
	); err != nil {
		return err
	}

	return h.rides.MarkChecked(ctx, ride.RideID, "weather_checked_at", time.Now())
}

func (h *Handlers) stationPair(ctx context.Context, originCode, destinationCode string) (*stations.Station, *stations.Station, error) {
	origin, err := h.stations.Find(ctx, originCode)
	if err != nil {
		return nil, nil, err
	}

	destination, err := h.stations.Find(ctx, destinationCode)
	if err != nil {
		return nil, nil, err
	}

	return origin, destination, nil
}
