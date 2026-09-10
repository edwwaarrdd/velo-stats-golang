package testsupport

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"database/sql"

	"velostats/internal/database"
	"velostats/internal/rides"
	"velostats/internal/routing"
	"velostats/internal/stations"
	"velostats/internal/weather"
)

func NewDatabase(t *testing.T) *sql.DB {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

func Time(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}

	return parsed
}

func Ride(t *testing.T, db *sql.DB, ride rides.Ride) rides.Ride {
	t.Helper()

	if ride.RideID == 0 {
		ride.RideID = nextRideID()
	}

	if ride.AccountID == 0 {
		ride.AccountID = 123
	}

	if ride.Status == "" {
		ride.Status = "Completed"
	}

	if ride.BikeNumber == "" {
		ride.BikeNumber = "5097"
	}

	if ride.OriginStationCode == "" {
		ride.OriginStationCode = "021"
	}

	if ride.OriginStation == "" {
		ride.OriginStation = "021- Driekoningen"
	}

	if ride.OriginSlotID == "" {
		ride.OriginSlotID = "15"
	}

	if ride.DestinationStationCode == "" {
		ride.DestinationStationCode = "041"
	}

	if ride.DestinationStation == "" {
		ride.DestinationStation = "041- Van Eyck"
	}

	if ride.DestinationSlotID == "" {
		ride.DestinationSlotID = "23"
	}

	if ride.CheckoutTime.IsZero() {
		ride.CheckoutTime = Time(t, "2026-09-06 08:57:02")
	}

	if ride.Duration == 0 {
		ride.Duration = 8
	}

	if ride.CheckinTime.IsZero() {
		ride.CheckinTime = ride.CheckoutTime.Add(time.Duration(ride.Duration) * time.Minute)
	}

	if _, err := rides.NewRepository(db).Save(context.Background(), ride); err != nil {
		t.Fatalf("store ride %d: %v", ride.RideID, err)
	}

	return ride
}

func Station(t *testing.T, db *sql.DB, station stations.Station) stations.Station {
	t.Helper()

	if station.Name == "" {
		station.Name = "Station " + station.StationID
	}

	if station.ShortName == "" {
		station.ShortName = station.StationID
	}

	if station.Lat == 0 {
		station.Lat = 51.19548
	}

	if station.Lon == 0 {
		station.Lon = 4.41919
	}

	if station.Address == "" {
		station.Address = "Driekoningenstraat 1"
	}

	if station.PostCode == "" {
		station.PostCode = "2600"
	}

	if station.RentalMethods == nil {
		station.RentalMethods = []string{"KEY"}
	}

	if station.Capacity == 0 {
		station.Capacity = 28
	}

	if _, err := stations.NewRepository(db).Save(context.Background(), station); err != nil {
		t.Fatalf("store station %s: %v", station.StationID, err)
	}

	return station
}

func BikeRouteBetween(t *testing.T, db *sql.DB, originCode, destinationCode string, distanceMeters, durationSeconds float64) {
	t.Helper()

	CachedRoute(t, db, originCode, destinationCode, routing.ModeBike, distanceMeters, durationSeconds)
}

func CachedRoute(t *testing.T, db *sql.DB, originCode, destinationCode string, mode routing.TravelMode, distanceMeters, durationSeconds float64) {
	t.Helper()

	repository := stations.NewRepository(db)

	for _, code := range []string{originCode, destinationCode} {
		existing, err := repository.Find(context.Background(), code)
		if err != nil {
			t.Fatalf("look up station %s: %v", code, err)
		}

		if existing == nil {
			Station(t, db, stations.Station{StationID: code})
		}
	}

	if err := routing.NewRepository(db).Save(
		context.Background(), originCode, destinationCode, mode,
		routing.Route{DistanceMeters: distanceMeters, DurationSeconds: durationSeconds},
	); err != nil {
		t.Fatalf("cache route %s -> %s: %v", originCode, destinationCode, err)
	}
}

func WeatherRecord(t *testing.T, db *sql.DB, rideID int64, observation weather.Observation) {
	t.Helper()

	if err := weather.NewRepository(db).Save(context.Background(), rideID, observation); err != nil {
		t.Fatalf("cache weather for ride %d: %v", rideID, err)
	}
}

var lastRideID int64 = 1000

func nextRideID() int64 {
	lastRideID++

	return lastRideID
}
