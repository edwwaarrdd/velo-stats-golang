// Package rides holds the ride history: the model, its storage, the export it
// is loaded from, the statistics derived from it and the endpoints serving it.
package rides

import (
	"time"

	"velostats/internal/weather"
)

// Ride is a single trip between two stations.
type Ride struct {
	RideID                 int64
	AccountID              int64
	Status                 string
	Duration               int
	BikeNumber             string
	OriginStationCode      string
	OriginStation          string
	OriginSlotID           string
	CheckoutTime           time.Time
	DestinationStationCode string
	DestinationStation     string
	DestinationSlotID      string
	CheckinTime            time.Time
	DistanceCheckedAt      *time.Time
	WeatherCheckedAt       *time.Time
}

// ListedRide is a ride enriched with everything the listing endpoint reports:
// the cached route between its stations and the weather it was ridden in.
type ListedRide struct {
	Ride

	DistanceMeters          *float64
	ExpectedDurationSeconds *float64
	Weather                 *weather.Observation
}

// ActualDurationSeconds is the ride time to the second, since Duration is only
// stored in whole minutes.
func (r Ride) ActualDurationSeconds() float64 {
	return r.CheckinTime.Sub(r.CheckoutTime).Seconds()
}
