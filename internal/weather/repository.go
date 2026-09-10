package weather

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"velostats/internal/database"
	"velostats/internal/support"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindForRide(ctx context.Context, rideID int64) (*Observation, error) {
	var (
		observation Observation
		observedAt  string
	)

	err := r.db.QueryRowContext(ctx,
		`SELECT temperature_c, apparent_temperature_c, precipitation_mm, rain_mm, snowfall_cm,
			cloud_cover_percent, wind_speed_kmh, wind_gusts_kmh, wind_direction_degrees,
			relative_humidity_percent, weather_code, observed_at
		FROM weather_records WHERE ride_id = ?`,
		rideID,
	).Scan(
		&observation.TemperatureC, &observation.ApparentTemperatureC, &observation.PrecipitationMm,
		&observation.RainMm, &observation.SnowfallCm, &observation.CloudCoverPercent,
		&observation.WindSpeedKmh, &observation.WindGustsKmh, &observation.WindDirectionDegrees,
		&observation.RelativeHumidityPercent, &observation.WeatherCode, &observedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("find cached weather for ride %d: %w", rideID, err)
	}

	if observation.ObservedAt, err = database.ParseDateTime(observedAt); err != nil {
		return nil, fmt.Errorf("read observed_at for ride %d: %w", rideID, err)
	}

	return &observation, nil
}

func (r *Repository) Save(ctx context.Context, rideID int64, observation Observation) error {
	now := support.DatabaseDateTime(time.Now())

	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO weather_records (
			ride_id, temperature_c, apparent_temperature_c, precipitation_mm, rain_mm, snowfall_cm,
			cloud_cover_percent, wind_speed_kmh, wind_gusts_kmh, wind_direction_degrees,
			relative_humidity_percent, weather_code, observed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (ride_id) DO UPDATE SET
			temperature_c = excluded.temperature_c,
			apparent_temperature_c = excluded.apparent_temperature_c,
			precipitation_mm = excluded.precipitation_mm,
			rain_mm = excluded.rain_mm,
			snowfall_cm = excluded.snowfall_cm,
			cloud_cover_percent = excluded.cloud_cover_percent,
			wind_speed_kmh = excluded.wind_speed_kmh,
			wind_gusts_kmh = excluded.wind_gusts_kmh,
			wind_direction_degrees = excluded.wind_direction_degrees,
			relative_humidity_percent = excluded.relative_humidity_percent,
			weather_code = excluded.weather_code,
			observed_at = excluded.observed_at,
			updated_at = excluded.updated_at`,
		rideID, observation.TemperatureC, observation.ApparentTemperatureC, observation.PrecipitationMm,
		observation.RainMm, observation.SnowfallCm, observation.CloudCoverPercent,
		observation.WindSpeedKmh, observation.WindGustsKmh, observation.WindDirectionDegrees,
		observation.RelativeHumidityPercent, observation.WeatherCode,
		support.DatabaseDateTime(observation.ObservedAt), now, now,
	); err != nil {
		return fmt.Errorf("cache weather for ride %d: %w", rideID, err)
	}

	return nil
}
