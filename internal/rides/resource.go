package rides

import (
	"velostats/internal/support"
	"velostats/internal/weather"
)

type Resource struct {
	RideID                    int64             `json:"ride_id"`
	AccountID                 int64             `json:"account_id"`
	Status                    string            `json:"status"`
	Duration                  int               `json:"duration"`
	BikeNumber                string            `json:"bike_number"`
	OriginStationCode         string            `json:"origin_station_code"`
	OriginStation             string            `json:"origin_station"`
	OriginSlotID              string            `json:"origin_slot_id"`
	CheckoutTime              string            `json:"checkout_time"`
	DestinationStationCode    string            `json:"destination_station_code"`
	DestinationStation        string            `json:"destination_station"`
	DestinationSlotID         string            `json:"destination_slot_id"`
	CheckinTime               string            `json:"checkin_time"`
	DistanceMeters            support.NullFloat `json:"distance_meters"`
	SpeedKmh                  support.NullFloat `json:"speed_kmh"`
	ExpectedDurationSeconds   support.NullFloat `json:"expected_duration_seconds"`
	ActualDurationSeconds     support.NullFloat `json:"actual_duration_seconds"`
	DurationVsExpectedSeconds support.NullFloat `json:"duration_vs_expected_seconds"`
	Weather                   *weather.Resource `json:"weather"`
}

func NewResource(ride ListedRide) Resource {
	actualDurationSeconds := support.Money(ride.ActualDurationSeconds())

	resource := Resource{
		RideID:                    ride.RideID,
		AccountID:                 ride.AccountID,
		Status:                    ride.Status,
		Duration:                  ride.Duration,
		BikeNumber:                ride.BikeNumber,
		OriginStationCode:         ride.OriginStationCode,
		OriginStation:             ride.OriginStation,
		OriginSlotID:              ride.OriginSlotID,
		CheckoutTime:              support.APIDateTime(ride.CheckoutTime),
		DestinationStationCode:    ride.DestinationStationCode,
		DestinationStation:        ride.DestinationStation,
		DestinationSlotID:         ride.DestinationSlotID,
		CheckinTime:               support.APIDateTime(ride.CheckinTime),
		DistanceMeters:            support.FloatPtr(ride.DistanceMeters),
		SpeedKmh:                  speedKmh(ride, actualDurationSeconds),
		ExpectedDurationSeconds:   support.FloatPtr(support.MoneyPtr(ride.ExpectedDurationSeconds)),
		ActualDurationSeconds:     support.FloatValue(actualDurationSeconds),
		DurationVsExpectedSeconds: durationVsExpectedSeconds(ride, actualDurationSeconds),
	}

	if ride.Weather != nil {
		rendered := weather.NewResource(*ride.Weather)
		resource.Weather = &rendered
	}

	return resource
}

// This divides by the exact ride time rather than by the Duration field, which
// truncates to whole minutes and so overstates the speed.
func speedKmh(ride ListedRide, actualDurationSeconds float64) support.NullFloat {
	if ride.DistanceMeters == nil || actualDurationSeconds <= 0 {
		return support.NullFloat{}
	}

	return support.FloatValue(support.Money((*ride.DistanceMeters / 1000) / (actualDurationSeconds / 3600)))
}

func durationVsExpectedSeconds(ride ListedRide, actualDurationSeconds float64) support.NullFloat {
	if ride.ExpectedDurationSeconds == nil {
		return support.NullFloat{}
	}

	return support.FloatValue(support.Money(actualDurationSeconds - *ride.ExpectedDurationSeconds))
}
