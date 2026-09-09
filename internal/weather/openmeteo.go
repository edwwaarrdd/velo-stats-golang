package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"velostats/internal/support"
)

// Service fetches the weather observed at a location for the hour of a time.
type Service interface {
	GetWeather(ctx context.Context, location support.Coordinate, at time.Time) (Observation, error)
}

// HourlyVariables are the biking-relevant hourly variables requested from the
// Open-Meteo archive.
var HourlyVariables = []string{
	"temperature_2m",
	"apparent_temperature",
	"precipitation",
	"rain",
	"snowfall",
	"cloud_cover",
	"wind_speed_10m",
	"wind_gusts_10m",
	"wind_direction_10m",
	"relative_humidity_2m",
	"weather_code",
}

// OpenMeteoService reads observations from the free Open-Meteo archive.
type OpenMeteoService struct {
	client     *http.Client
	archiveURL string
}

// NewOpenMeteoService builds the Open-Meteo backed weather service.
func NewOpenMeteoService(client *http.Client, archiveURL string) *OpenMeteoService {
	return &OpenMeteoService{client: client, archiveURL: archiveURL}
}

type archiveResponse struct {
	Reason string `json:"reason"`
	Hourly *struct {
		Time                []string  `json:"time"`
		Temperature         []float64 `json:"temperature_2m"`
		ApparentTemperature []float64 `json:"apparent_temperature"`
		Precipitation       []float64 `json:"precipitation"`
		Rain                []float64 `json:"rain"`
		Snowfall            []float64 `json:"snowfall"`
		CloudCover          []float64 `json:"cloud_cover"`
		WindSpeed           []float64 `json:"wind_speed_10m"`
		WindGusts           []float64 `json:"wind_gusts_10m"`
		WindDirection       []float64 `json:"wind_direction_10m"`
		RelativeHumidity    []float64 `json:"relative_humidity_2m"`
		WeatherCode         []int     `json:"weather_code"`
	} `json:"hourly"`
}

// GetWeather returns the observation for the hour of the given time.
func (s *OpenMeteoService) GetWeather(ctx context.Context, location support.Coordinate, at time.Time) (Observation, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	at = at.UTC()
	date := at.Format(support.APIDateFormat)

	query := url.Values{}
	query.Set("latitude", strconv.FormatFloat(location.Lat, 'f', -1, 64))
	query.Set("longitude", strconv.FormatFloat(location.Lon, 'f', -1, 64))
	query.Set("start_date", date)
	query.Set("end_date", date)
	query.Set("hourly", strings.Join(HourlyVariables, ","))
	query.Set("timezone", "UTC")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.archiveURL+"?"+query.Encode(), nil)
	if err != nil {
		return Observation{}, fmt.Errorf("build Open-Meteo request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return Observation{}, fmt.Errorf("Open-Meteo request failed: %w", err)
	}
	defer response.Body.Close()

	var payload archiveResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Observation{}, fmt.Errorf("decode Open-Meteo response: %w", err)
	}

	if payload.Hourly == nil {
		return Observation{}, fmt.Errorf("Open-Meteo request failed: %s", firstNonEmpty(payload.Reason, "unknown error"))
	}

	targetHour := at.Truncate(time.Hour).Format("2006-01-02T15:00")

	index := -1
	for position, hour := range payload.Hourly.Time {
		if hour == targetHour {
			index = position

			break
		}
	}

	if index < 0 {
		return Observation{}, fmt.Errorf("Open-Meteo response has no observation for %s", targetHour)
	}

	observedAt, err := time.ParseInLocation("2006-01-02T15:04", payload.Hourly.Time[index], time.UTC)
	if err != nil {
		return Observation{}, fmt.Errorf("parse Open-Meteo observation time: %w", err)
	}

	hourly := payload.Hourly

	return Observation{
		TemperatureC:            hourly.Temperature[index],
		ApparentTemperatureC:    hourly.ApparentTemperature[index],
		PrecipitationMm:         hourly.Precipitation[index],
		RainMm:                  hourly.Rain[index],
		SnowfallCm:              hourly.Snowfall[index],
		CloudCoverPercent:       hourly.CloudCover[index],
		WindSpeedKmh:            hourly.WindSpeed[index],
		WindGustsKmh:            hourly.WindGusts[index],
		WindDirectionDegrees:    hourly.WindDirection[index],
		RelativeHumidityPercent: hourly.RelativeHumidity[index],
		WeatherCode:             hourly.WeatherCode[index],
		ObservedAt:              observedAt,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
