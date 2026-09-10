package weather_test

import (
	"context"
	"testing"
	"time"

	"velostats/internal/rides"
	"velostats/internal/support"
	"velostats/internal/testsupport"
	"velostats/internal/weather"
)

type countingService struct {
	observation weather.Observation
	calls       int
}

func (s *countingService) GetWeather(context.Context, support.Coordinate, time.Time) (weather.Observation, error) {
	s.calls++

	return s.observation, nil
}

func observation(t *testing.T, temperature float64) weather.Observation {
	t.Helper()

	return weather.Observation{
		TemperatureC:            temperature,
		ApparentTemperatureC:    17.1,
		CloudCoverPercent:       42.0,
		WindSpeedKmh:            11.2,
		WindGustsKmh:            24.5,
		WindDirectionDegrees:    210.0,
		RelativeHumidityPercent: 68.0,
		WeatherCode:             3,
		ObservedAt:              at(t, "2026-09-06 09:00:00"),
	}
}

func TestCachedRideWeatherOnlyFetchesARidesWeatherOnce(t *testing.T) {
	db := testsupport.NewDatabase(t)
	ride := testsupport.Ride(t, db, rides.Ride{})

	upstream := &countingService{observation: observation(t, 18.0)}
	service := weather.NewCachedRideWeatherService(weather.NewRepository(db), upstream)

	for range 3 {
		got, err := service.GetWeather(context.Background(), ride.RideID, ride.CheckinTime, support.Coordinate{}, false)
		if err != nil {
			t.Fatalf("GetWeather: %v", err)
		}

		if got.TemperatureC != 18.0 {
			t.Errorf("temperature = %v, want 18", got.TemperatureC)
		}
	}

	if upstream.calls != 1 {
		t.Errorf("the archive was called %d times, want 1", upstream.calls)
	}
}

func TestCachedRideWeatherRefetchesWhenForced(t *testing.T) {
	db := testsupport.NewDatabase(t)
	ride := testsupport.Ride(t, db, rides.Ride{})

	upstream := &countingService{observation: observation(t, 18.0)}
	service := weather.NewCachedRideWeatherService(weather.NewRepository(db), upstream)

	if _, err := service.GetWeather(context.Background(), ride.RideID, ride.CheckinTime, support.Coordinate{}, false); err != nil {
		t.Fatalf("GetWeather: %v", err)
	}

	upstream.observation = observation(t, 21.5)

	got, err := service.GetWeather(context.Background(), ride.RideID, ride.CheckinTime, support.Coordinate{}, true)
	if err != nil {
		t.Fatalf("GetWeather: %v", err)
	}

	if got.TemperatureC != 21.5 {
		t.Errorf("temperature = %v, want the refetched 21.5", got.TemperatureC)
	}

	if upstream.calls != 2 {
		t.Errorf("the archive was called %d times, want 2", upstream.calls)
	}

	cached, err := weather.NewRepository(db).FindForRide(context.Background(), ride.RideID)
	if err != nil {
		t.Fatalf("FindForRide: %v", err)
	}

	if cached == nil || cached.TemperatureC != 21.5 {
		t.Errorf("the cache should hold the refetched observation, got %+v", cached)
	}
}
