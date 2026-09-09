package jobs_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"velostats/internal/jobs"
	"velostats/internal/rides"
	"velostats/internal/routing"
	"velostats/internal/stations"
	"velostats/internal/support"
	"velostats/internal/testsupport"
	"velostats/internal/weather"

	"database/sql"
)

type stubRouter struct {
	route routing.Route
	calls int
}

func (s *stubRouter) GetRoute(context.Context, support.Coordinate, support.Coordinate, routing.TravelMode) (routing.Route, error) {
	s.calls++

	return s.route, nil
}

type stubArchive struct {
	observation weather.Observation
	calls       int
}

func (s *stubArchive) GetWeather(context.Context, support.Coordinate, time.Time) (weather.Observation, error) {
	s.calls++

	return s.observation, nil
}

// newHandlers wires the job handlers over a test database and stub upstreams.
func newHandlers(t *testing.T, db *sql.DB, router *stubRouter, archive *stubArchive) *jobs.Handlers {
	t.Helper()

	return jobs.NewHandlers(
		rides.NewRepository(db),
		stations.NewRepository(db),
		routing.NewCachedStationRouteService(routing.NewRepository(db), router),
		weather.NewCachedRideWeatherService(weather.NewRepository(db), archive),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func payload(t *testing.T, value any) json.RawMessage {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	return encoded
}

func TestCheckRideDistanceCachesTheRouteAndMarksTheRideChecked(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Station(t, db, stations.Station{StationID: "021"})
	testsupport.Station(t, db, stations.Station{StationID: "041"})
	ride := testsupport.Ride(t, db, rides.Ride{OriginStationCode: "021", DestinationStationCode: "041"})

	router := &stubRouter{route: routing.Route{DistanceMeters: 1502.3, DurationSeconds: 361.7}}

	if err := newHandlers(t, db, router, &stubArchive{}).CheckRideDistance(
		context.Background(),
		payload(t, jobs.CheckRideDistancePayload{RideID: ride.RideID}),
	); err != nil {
		t.Fatalf("CheckRideDistance: %v", err)
	}

	cached, err := routing.NewRepository(db).Find(context.Background(), "021", "041", routing.ModeBike)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	if cached == nil || cached.DistanceMeters != 1502.3 {
		t.Fatalf("the route should be cached, got %+v", cached)
	}

	stored, err := rides.NewRepository(db).Find(context.Background(), ride.RideID)
	if err != nil {
		t.Fatalf("Find ride: %v", err)
	}

	if stored.DistanceCheckedAt == nil {
		t.Error("the ride should be marked as distance checked")
	}
}

func TestCheckRideDistanceSkipsARideThatWasAlreadyChecked(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Station(t, db, stations.Station{StationID: "021"})
	testsupport.Station(t, db, stations.Station{StationID: "041"})

	checkedAt := testsupport.Time(t, "2026-01-01 00:00:00")
	ride := testsupport.Ride(t, db, rides.Ride{DistanceCheckedAt: &checkedAt})

	router := &stubRouter{}

	if err := newHandlers(t, db, router, &stubArchive{}).CheckRideDistance(
		context.Background(),
		payload(t, jobs.CheckRideDistancePayload{RideID: ride.RideID}),
	); err != nil {
		t.Fatalf("CheckRideDistance: %v", err)
	}

	if router.calls != 0 {
		t.Errorf("the router was called %d times, want none", router.calls)
	}
}

func TestCheckRideDistanceSkipsARideWithAnUnknownStation(t *testing.T) {
	db := testsupport.NewDatabase(t)

	ride := testsupport.Ride(t, db, rides.Ride{OriginStationCode: "021", DestinationStationCode: "999"})
	router := &stubRouter{}

	if err := newHandlers(t, db, router, &stubArchive{}).CheckRideDistance(
		context.Background(),
		payload(t, jobs.CheckRideDistancePayload{RideID: ride.RideID}),
	); err != nil {
		t.Fatalf("CheckRideDistance: %v", err)
	}

	if router.calls != 0 {
		t.Errorf("the router was called %d times, want none", router.calls)
	}

	stored, err := rides.NewRepository(db).Find(context.Background(), ride.RideID)
	if err != nil {
		t.Fatalf("Find ride: %v", err)
	}

	if stored.DistanceCheckedAt != nil {
		t.Error("a ride with an unknown station should stay unchecked")
	}
}

func TestCheckRideWeatherCachesTheObservationAndMarksTheRideChecked(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Station(t, db, stations.Station{StationID: "021"})
	ride := testsupport.Ride(t, db, rides.Ride{OriginStationCode: "021"})

	archive := &stubArchive{observation: weather.Observation{
		TemperatureC: 18.0,
		WeatherCode:  3,
		ObservedAt:   testsupport.Time(t, "2026-09-06 09:00:00"),
	}}

	if err := newHandlers(t, db, &stubRouter{}, archive).CheckRideWeather(
		context.Background(),
		payload(t, jobs.CheckRideWeatherPayload{RideID: ride.RideID}),
	); err != nil {
		t.Fatalf("CheckRideWeather: %v", err)
	}

	cached, err := weather.NewRepository(db).FindForRide(context.Background(), ride.RideID)
	if err != nil {
		t.Fatalf("FindForRide: %v", err)
	}

	if cached == nil || cached.TemperatureC != 18.0 {
		t.Fatalf("the observation should be cached, got %+v", cached)
	}

	stored, err := rides.NewRepository(db).Find(context.Background(), ride.RideID)
	if err != nil {
		t.Fatalf("Find ride: %v", err)
	}

	if stored.WeatherCheckedAt == nil {
		t.Error("the ride should be marked as weather checked")
	}
}

func TestCheckRideWeatherRefetchesAnAlreadyCheckedRideWhenForced(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Station(t, db, stations.Station{StationID: "021"})

	checkedAt := testsupport.Time(t, "2026-01-01 00:00:00")
	ride := testsupport.Ride(t, db, rides.Ride{OriginStationCode: "021", WeatherCheckedAt: &checkedAt})

	archive := &stubArchive{observation: weather.Observation{
		TemperatureC: 21.5,
		ObservedAt:   testsupport.Time(t, "2026-09-06 09:00:00"),
	}}

	handlers := newHandlers(t, db, &stubRouter{}, archive)

	if err := handlers.CheckRideWeather(
		context.Background(),
		payload(t, jobs.CheckRideWeatherPayload{RideID: ride.RideID}),
	); err != nil {
		t.Fatalf("CheckRideWeather: %v", err)
	}

	if archive.calls != 0 {
		t.Errorf("an already checked ride was fetched %d times, want none", archive.calls)
	}

	if err := handlers.CheckRideWeather(
		context.Background(),
		payload(t, jobs.CheckRideWeatherPayload{RideID: ride.RideID, Force: true}),
	); err != nil {
		t.Fatalf("CheckRideWeather forced: %v", err)
	}

	if archive.calls != 1 {
		t.Errorf("a forced check fetched %d times, want 1", archive.calls)
	}
}
