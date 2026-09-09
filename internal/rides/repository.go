package rides

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"velostats/internal/database"
	"velostats/internal/routing"
	"velostats/internal/support"
	"velostats/internal/weather"
)

// Repository stores and reads rides.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a repository over the given database.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// cachedRouteColumn is the correlated subquery that resolves a ride's cycling
// distance or expected ride time from the cached route between its origin and
// destination stations.
func cachedRouteColumn(column string) string {
	return `(SELECT sr.` + column + ` FROM station_routes sr
		WHERE sr.origin_station_id = rides.origin_station_code
			AND sr.destination_station_id = rides.destination_station_code
			AND sr.mode = '` + string(routing.ModeBike) + `'
		LIMIT 1)`
}

const rideColumns = `rides.ride_id, rides.account_id, rides.status, rides.duration, rides.bike_number,
	rides.origin_station_code, rides.origin_station, rides.origin_slot_id, rides.checkout_time,
	rides.destination_station_code, rides.destination_station, rides.destination_slot_id, rides.checkin_time,
	rides.distance_checked_at, rides.weather_checked_at`

// All returns every ride, most recent first, with its cached distance, expected
// ride time and weather.
func (r *Repository) All(ctx context.Context) ([]ListedRide, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+rideColumns+`,
			`+cachedRouteColumn("distance_meters")+` AS distance_meters,
			`+cachedRouteColumn("duration_seconds")+` AS expected_duration_seconds,
			w.temperature_c, w.apparent_temperature_c, w.precipitation_mm, w.rain_mm, w.snowfall_cm,
			w.cloud_cover_percent, w.wind_speed_kmh, w.wind_gusts_kmh, w.wind_direction_degrees,
			w.relative_humidity_percent, w.weather_code, w.observed_at
		FROM rides
		LEFT JOIN weather_records w ON w.ride_id = rides.ride_id
		ORDER BY rides.checkout_time DESC, rides.ride_id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list rides: %w", err)
	}
	defer rows.Close()

	listed := []ListedRide{}

	for rows.Next() {
		ride, err := scanListedRide(rows)
		if err != nil {
			return nil, err
		}

		listed = append(listed, ride)
	}

	return listed, rows.Err()
}

// Find returns the ride with the given id, or nil when it is unknown.
func (r *Repository) Find(ctx context.Context, rideID int64) (*Ride, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+rideColumns+` FROM rides WHERE ride_id = ?`, rideID)

	ride, err := scanRide(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &ride, nil
}

// Save inserts the ride, or updates it when it already exists. It reports
// whether the ride was newly created.
func (r *Repository) Save(ctx context.Context, ride Ride) (bool, error) {
	now := support.DatabaseDateTime(time.Now())

	result, err := r.db.ExecContext(ctx, `UPDATE rides SET
			account_id = ?, status = ?, duration = ?, bike_number = ?,
			origin_station_code = ?, origin_station = ?, origin_slot_id = ?, checkout_time = ?,
			destination_station_code = ?, destination_station = ?, destination_slot_id = ?, checkin_time = ?,
			updated_at = ?
		WHERE ride_id = ?`,
		ride.AccountID, ride.Status, ride.Duration, ride.BikeNumber,
		ride.OriginStationCode, ride.OriginStation, ride.OriginSlotID, support.DatabaseDateTime(ride.CheckoutTime),
		ride.DestinationStationCode, ride.DestinationStation, ride.DestinationSlotID, support.DatabaseDateTime(ride.CheckinTime),
		now, ride.RideID,
	)
	if err != nil {
		return false, fmt.Errorf("update ride %d: %w", ride.RideID, err)
	}

	if updated, err := result.RowsAffected(); err == nil && updated > 0 {
		return false, nil
	}

	if _, err := r.db.ExecContext(ctx, `INSERT INTO rides (
			ride_id, account_id, status, duration, bike_number,
			origin_station_code, origin_station, origin_slot_id, checkout_time,
			destination_station_code, destination_station, destination_slot_id, checkin_time,
			distance_checked_at, weather_checked_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ride.RideID, ride.AccountID, ride.Status, ride.Duration, ride.BikeNumber,
		ride.OriginStationCode, ride.OriginStation, ride.OriginSlotID, support.DatabaseDateTime(ride.CheckoutTime),
		ride.DestinationStationCode, ride.DestinationStation, ride.DestinationSlotID, support.DatabaseDateTime(ride.CheckinTime),
		database.NullTimeValue(ride.DistanceCheckedAt), database.NullTimeValue(ride.WeatherCheckedAt), now, now,
	); err != nil {
		return false, fmt.Errorf("insert ride %d: %w", ride.RideID, err)
	}

	return true, nil
}

// IDs returns the id of every ride, optionally only those whose given check
// column has not run yet.
func (r *Repository) IDs(ctx context.Context, uncheckedColumn string) ([]int64, error) {
	query := `SELECT ride_id FROM rides`
	if uncheckedColumn != "" {
		query += ` WHERE ` + uncheckedColumn + ` IS NULL`
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list ride ids: %w", err)
	}
	defer rows.Close()

	ids := []int64{}

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("read ride id: %w", err)
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// MarkChecked records that a background check has run for a ride.
func (r *Repository) MarkChecked(ctx context.Context, rideID int64, column string, at time.Time) error {
	if _, err := r.db.ExecContext(ctx,
		`UPDATE rides SET `+column+` = ?, updated_at = ? WHERE ride_id = ?`,
		support.DatabaseDateTime(at), support.DatabaseDateTime(time.Now()), rideID,
	); err != nil {
		return fmt.Errorf("mark ride %d as checked: %w", rideID, err)
	}

	return nil
}

// CheckoutTimes returns the check-out time of every ride, in UTC.
func (r *Repository) CheckoutTimes(ctx context.Context) ([]time.Time, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT checkout_time FROM rides`)
	if err != nil {
		return nil, fmt.Errorf("list check-out times: %w", err)
	}
	defer rows.Close()

	times := []time.Time{}

	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("read check-out time: %w", err)
		}

		parsed, err := database.ParseDateTime(raw)
		if err != nil {
			return nil, err
		}

		times = append(times, parsed)
	}

	return times, rows.Err()
}

func scanRide(source interface{ Scan(...any) error }) (Ride, error) {
	var (
		ride              Ride
		checkoutTime      string
		checkinTime       string
		distanceCheckedAt sql.NullString
		weatherCheckedAt  sql.NullString
	)

	if err := source.Scan(
		&ride.RideID, &ride.AccountID, &ride.Status, &ride.Duration, &ride.BikeNumber,
		&ride.OriginStationCode, &ride.OriginStation, &ride.OriginSlotID, &checkoutTime,
		&ride.DestinationStationCode, &ride.DestinationStation, &ride.DestinationSlotID, &checkinTime,
		&distanceCheckedAt, &weatherCheckedAt,
	); err != nil {
		return Ride{}, err
	}

	return hydrateTimes(ride, checkoutTime, checkinTime, distanceCheckedAt, weatherCheckedAt)
}

func scanListedRide(rows *sql.Rows) (ListedRide, error) {
	var (
		ride              Ride
		checkoutTime      string
		checkinTime       string
		distanceCheckedAt sql.NullString
		weatherCheckedAt  sql.NullString

		distanceMeters          sql.NullFloat64
		expectedDurationSeconds sql.NullFloat64

		temperature      sql.NullFloat64
		apparent         sql.NullFloat64
		precipitation    sql.NullFloat64
		rain             sql.NullFloat64
		snowfall         sql.NullFloat64
		cloudCover       sql.NullFloat64
		windSpeed        sql.NullFloat64
		windGusts        sql.NullFloat64
		windDirection    sql.NullFloat64
		relativeHumidity sql.NullFloat64
		weatherCode      sql.NullInt64
		observedAt       sql.NullString
	)

	if err := rows.Scan(
		&ride.RideID, &ride.AccountID, &ride.Status, &ride.Duration, &ride.BikeNumber,
		&ride.OriginStationCode, &ride.OriginStation, &ride.OriginSlotID, &checkoutTime,
		&ride.DestinationStationCode, &ride.DestinationStation, &ride.DestinationSlotID, &checkinTime,
		&distanceCheckedAt, &weatherCheckedAt,
		&distanceMeters, &expectedDurationSeconds,
		&temperature, &apparent, &precipitation, &rain, &snowfall, &cloudCover,
		&windSpeed, &windGusts, &windDirection, &relativeHumidity, &weatherCode, &observedAt,
	); err != nil {
		return ListedRide{}, fmt.Errorf("read ride: %w", err)
	}

	hydrated, err := hydrateTimes(ride, checkoutTime, checkinTime, distanceCheckedAt, weatherCheckedAt)
	if err != nil {
		return ListedRide{}, err
	}

	listed := ListedRide{
		Ride:                    hydrated,
		DistanceMeters:          database.FloatPointer(distanceMeters),
		ExpectedDurationSeconds: database.FloatPointer(expectedDurationSeconds),
	}

	if observedAt.Valid {
		observed, err := database.ParseDateTime(observedAt.String)
		if err != nil {
			return ListedRide{}, err
		}

		listed.Weather = &weather.Observation{
			TemperatureC:            temperature.Float64,
			ApparentTemperatureC:    apparent.Float64,
			PrecipitationMm:         precipitation.Float64,
			RainMm:                  rain.Float64,
			SnowfallCm:              snowfall.Float64,
			CloudCoverPercent:       cloudCover.Float64,
			WindSpeedKmh:            windSpeed.Float64,
			WindGustsKmh:            windGusts.Float64,
			WindDirectionDegrees:    windDirection.Float64,
			RelativeHumidityPercent: relativeHumidity.Float64,
			WeatherCode:             int(weatherCode.Int64),
			ObservedAt:              observed,
		}
	}

	return listed, nil
}

func hydrateTimes(ride Ride, checkoutTime, checkinTime string, distanceCheckedAt, weatherCheckedAt sql.NullString) (Ride, error) {
	var err error

	if ride.CheckoutTime, err = database.ParseDateTime(checkoutTime); err != nil {
		return Ride{}, fmt.Errorf("read checkout_time for ride %d: %w", ride.RideID, err)
	}

	if ride.CheckinTime, err = database.ParseDateTime(checkinTime); err != nil {
		return Ride{}, fmt.Errorf("read checkin_time for ride %d: %w", ride.RideID, err)
	}

	if ride.DistanceCheckedAt, err = database.NullTime(distanceCheckedAt); err != nil {
		return Ride{}, fmt.Errorf("read distance_checked_at for ride %d: %w", ride.RideID, err)
	}

	if ride.WeatherCheckedAt, err = database.NullTime(weatherCheckedAt); err != nil {
		return Ride{}, fmt.Errorf("read weather_checked_at for ride %d: %w", ride.RideID, err)
	}

	return ride, nil
}
