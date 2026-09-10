package rides

import (
	"time"

	"velostats/internal/weather"
)

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

type ListedRide struct {
	Ride

	DistanceMeters          *float64
	ExpectedDurationSeconds *float64
	Weather                 *weather.Observation
}

// Duration is only stored in whole minutes, so this recomputes the ride time to
// the second.
func (r Ride) ActualDurationSeconds() float64 {
	return r.CheckinTime.Sub(r.CheckoutTime).Seconds()
}
