package routing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"velostats/internal/support"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Find(ctx context.Context, originID, destinationID string, mode TravelMode) (*Route, error) {
	var route Route

	err := r.db.QueryRowContext(ctx,
		`SELECT distance_meters, duration_seconds FROM station_routes
		WHERE origin_station_id = ? AND destination_station_id = ? AND mode = ?
		LIMIT 1`,
		originID, destinationID, string(mode),
	).Scan(&route.DistanceMeters, &route.DurationSeconds)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("find cached route %s -> %s: %w", originID, destinationID, err)
	}

	return &route, nil
}

func (r *Repository) Save(ctx context.Context, originID, destinationID string, mode TravelMode, route Route) error {
	now := support.DatabaseDateTime(time.Now())

	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO station_routes (
			origin_station_id, destination_station_id, mode, distance_meters, duration_seconds, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (origin_station_id, destination_station_id, mode) DO UPDATE SET
			distance_meters = excluded.distance_meters,
			duration_seconds = excluded.duration_seconds,
			updated_at = excluded.updated_at`,
		originID, destinationID, string(mode), route.DistanceMeters, route.DurationSeconds, now, now,
	); err != nil {
		return fmt.Errorf("cache route %s -> %s: %w", originID, destinationID, err)
	}

	return nil
}
