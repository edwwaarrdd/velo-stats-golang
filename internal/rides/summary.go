package rides

import (
	"context"
	"database/sql"
	"fmt"

	"velostats/internal/database"
	"velostats/internal/support"
)

type Summary struct {
	TotalRides            int64             `json:"total_rides"`
	TotalDuration         support.NullInt   `json:"total_duration"`
	AverageDuration       support.NullFloat `json:"average_duration"`
	LongestRideDuration   support.NullInt   `json:"longest_ride_duration"`
	ShortestRideDuration  support.NullInt   `json:"shortest_ride_duration"`
	TotalDistanceMeters   support.NullFloat `json:"total_distance_meters"`
	AverageDistanceMeters support.NullFloat `json:"average_distance_meters"`
}

type SummaryCalculator struct {
	db *sql.DB
}

func NewSummaryCalculator(db *sql.DB) *SummaryCalculator {
	return &SummaryCalculator{db: db}
}

func (c *SummaryCalculator) Calculate(ctx context.Context) (Summary, error) {
	var (
		totalRides            int64
		totalDuration         sql.NullInt64
		averageDuration       sql.NullFloat64
		longestRideDuration   sql.NullInt64
		shortestRideDuration  sql.NullInt64
		totalDistanceMeters   sql.NullFloat64
		averageDistanceMeters sql.NullFloat64
	)

	if err := c.db.QueryRowContext(ctx, `SELECT
			COUNT(*), SUM(duration), AVG(duration), MAX(duration), MIN(duration),
			SUM(distance_meters), AVG(distance_meters)
		FROM (
			SELECT rides.duration, `+cachedRouteColumn("distance_meters")+` AS distance_meters FROM rides
		)`,
	).Scan(
		&totalRides, &totalDuration, &averageDuration, &longestRideDuration,
		&shortestRideDuration, &totalDistanceMeters, &averageDistanceMeters,
	); err != nil {
		return Summary{}, fmt.Errorf("summarise rides: %w", err)
	}

	return Summary{
		TotalRides:            totalRides,
		TotalDuration:         support.IntPtr(database.IntPointer(totalDuration)),
		AverageDuration:       support.FloatPtr(support.MoneyPtr(database.FloatPointer(averageDuration))),
		LongestRideDuration:   support.IntPtr(database.IntPointer(longestRideDuration)),
		ShortestRideDuration:  support.IntPtr(database.IntPointer(shortestRideDuration)),
		TotalDistanceMeters:   support.FloatPtr(database.FloatPointer(totalDistanceMeters)),
		AverageDistanceMeters: support.FloatPtr(support.MoneyPtr(database.FloatPointer(averageDistanceMeters))),
	}, nil
}
