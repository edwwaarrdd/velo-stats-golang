package weather_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"velostats/internal/support"
	"velostats/internal/weather"
)

const archiveBody = `{"hourly":{
	"time":["2026-09-06T08:00","2026-09-06T09:00"],
	"temperature_2m":[17.2,18.0],
	"apparent_temperature":[16.4,17.1],
	"precipitation":[0.0,0.2],
	"rain":[0.0,0.2],
	"snowfall":[0.0,0.0],
	"cloud_cover":[30.0,42.0],
	"wind_speed_10m":[9.8,11.2],
	"wind_gusts_10m":[20.1,24.5],
	"wind_direction_10m":[200.0,210.0],
	"relative_humidity_2m":[72.0,68.0],
	"weather_code":[1,3]
}}`

// fakeArchive serves a canned archive response and records the query it got.
func fakeArchive(t *testing.T, body string) (*weather.OpenMeteoService, *url.Values) {
	t.Helper()

	received := &url.Values{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*received = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return weather.NewOpenMeteoService(server.Client(), server.URL), received
}

func at(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}

	return parsed
}

func TestOpenMeteoReturnsTheObservationForTheHourOfTheGivenTime(t *testing.T) {
	service, _ := fakeArchive(t, archiveBody)

	observation, err := service.GetWeather(
		context.Background(),
		support.Coordinate{Lat: 51.19548, Lon: 4.41919},
		at(t, "2026-09-06 09:05:30"),
	)
	if err != nil {
		t.Fatalf("GetWeather: %v", err)
	}

	want := weather.Observation{
		TemperatureC:            18.0,
		ApparentTemperatureC:    17.1,
		PrecipitationMm:         0.2,
		RainMm:                  0.2,
		SnowfallCm:              0.0,
		CloudCoverPercent:       42.0,
		WindSpeedKmh:            11.2,
		WindGustsKmh:            24.5,
		WindDirectionDegrees:    210.0,
		RelativeHumidityPercent: 68.0,
		WeatherCode:             3,
		ObservedAt:              at(t, "2026-09-06 09:00:00"),
	}

	if observation != want {
		t.Errorf("observation = %+v, want %+v", observation, want)
	}
}

func TestOpenMeteoAsksForTheObservationDateInUTC(t *testing.T) {
	service, query := fakeArchive(t, archiveBody)

	if _, err := service.GetWeather(
		context.Background(),
		support.Coordinate{Lat: 51.19548, Lon: 4.41919},
		at(t, "2026-09-06 09:05:30"),
	); err != nil {
		t.Fatalf("GetWeather: %v", err)
	}

	if query.Get("start_date") != "2026-09-06" || query.Get("end_date") != "2026-09-06" {
		t.Errorf("date range = %s .. %s", query.Get("start_date"), query.Get("end_date"))
	}

	if query.Get("timezone") != "UTC" {
		t.Errorf("timezone = %s, want UTC", query.Get("timezone"))
	}

	if !strings.Contains(query.Get("hourly"), "apparent_temperature") {
		t.Errorf("hourly variables = %s", query.Get("hourly"))
	}
}

func TestOpenMeteoFailsWhenTheArchiveReturnsNoHourlyData(t *testing.T) {
	service, _ := fakeArchive(t, `{"reason":"out of range"}`)

	_, err := service.GetWeather(context.Background(), support.Coordinate{}, at(t, "2026-09-06 09:05:30"))
	if err == nil || !strings.Contains(err.Error(), "Open-Meteo request failed: out of range") {
		t.Errorf("error = %v, want the archive's reason", err)
	}
}

func TestOpenMeteoFailsWhenTheArchiveHasNoObservationForTheHour(t *testing.T) {
	service, _ := fakeArchive(t, archiveBody)

	_, err := service.GetWeather(context.Background(), support.Coordinate{}, at(t, "2026-09-06 23:05:30"))
	if err == nil || !strings.Contains(err.Error(), "no observation for 2026-09-06T23:00") {
		t.Errorf("error = %v, want a missing-hour failure", err)
	}
}
