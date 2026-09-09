package stations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"velostats/internal/support"
)

// Repository stores and reads stations.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a repository over the given database.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const selectColumns = `station_id, name, short_name, lat, lon, address, post_code, rental_methods, capacity`

// All returns every known station.
func (r *Repository) All(ctx context.Context) ([]Station, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+selectColumns+` FROM stations`)
	if err != nil {
		return nil, fmt.Errorf("list stations: %w", err)
	}
	defer rows.Close()

	found := []Station{}

	for rows.Next() {
		station, err := scan(rows)
		if err != nil {
			return nil, err
		}

		found = append(found, station)
	}

	return found, rows.Err()
}

// Find returns the station with the given id, or nil when it is unknown.
func (r *Repository) Find(ctx context.Context, stationID string) (*Station, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+selectColumns+` FROM stations WHERE station_id = ?`, stationID)

	station, err := scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &station, nil
}

// Save inserts the station, or updates it when it already exists. It reports
// whether the station was newly created.
func (r *Repository) Save(ctx context.Context, station Station) (bool, error) {
	rentalMethods, err := json.Marshal(station.RentalMethods)
	if err != nil {
		return false, fmt.Errorf("encode rental methods: %w", err)
	}

	now := support.DatabaseDateTime(time.Now())

	result, err := r.db.ExecContext(ctx, `UPDATE stations SET
			name = ?, short_name = ?, lat = ?, lon = ?, address = ?,
			post_code = ?, rental_methods = ?, capacity = ?, updated_at = ?
		WHERE station_id = ?`,
		station.Name, station.ShortName, station.Lat, station.Lon, station.Address,
		station.PostCode, string(rentalMethods), station.Capacity, now, station.StationID,
	)
	if err != nil {
		return false, fmt.Errorf("update station %s: %w", station.StationID, err)
	}

	if updated, err := result.RowsAffected(); err == nil && updated > 0 {
		return false, nil
	}

	if _, err := r.db.ExecContext(ctx, `INSERT INTO stations (
			`+selectColumns+`, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		station.StationID, station.Name, station.ShortName, station.Lat, station.Lon,
		station.Address, station.PostCode, string(rentalMethods), station.Capacity, now, now,
	); err != nil {
		return false, fmt.Errorf("insert station %s: %w", station.StationID, err)
	}

	return true, nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(destination ...any) error
}

func scan(source scanner) (Station, error) {
	var (
		station       Station
		rentalMethods string
	)

	if err := source.Scan(
		&station.StationID, &station.Name, &station.ShortName, &station.Lat, &station.Lon,
		&station.Address, &station.PostCode, &rentalMethods, &station.Capacity,
	); err != nil {
		return Station{}, err
	}

	if err := json.Unmarshal([]byte(rentalMethods), &station.RentalMethods); err != nil {
		return Station{}, fmt.Errorf("decode rental methods for station %s: %w", station.StationID, err)
	}

	return station, nil
}
