package httpapi_test

import (
	"encoding/json"
	"testing"

	"velostats/internal/rides"
	"velostats/internal/testsupport"
)

func TestSummaryReportsNoStatisticsWhenThereAreNoRides(t *testing.T) {
	_, body := get(t, testsupport.NewDatabase(t), "/rides/summary")

	assertJSON(t, body, `{
		"total_rides":0,
		"total_duration":null,
		"average_duration":null,
		"longest_ride_duration":null,
		"shortest_ride_duration":null,
		"total_distance_meters":null,
		"average_distance_meters":null
	}`)
}

func TestSummaryAggregatesDurationAndDistanceAcrossEveryRide(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 1500.0, 300.0)
	testsupport.BikeRouteBetween(t, db, "021", "076", 2500.0, 300.0)

	testsupport.Ride(t, db, rides.Ride{Duration: 10, OriginStationCode: "021", DestinationStationCode: "041"})
	testsupport.Ride(t, db, rides.Ride{Duration: 20, OriginStationCode: "021", DestinationStationCode: "076"})
	testsupport.Ride(t, db, rides.Ride{Duration: 15, OriginStationCode: "021", DestinationStationCode: "041"})

	_, body := get(t, db, "/rides/summary")

	assertJSON(t, body, `{
		"total_rides":3,
		"total_duration":45,
		"average_duration":15.0,
		"longest_ride_duration":20,
		"shortest_ride_duration":10,
		"total_distance_meters":5500.0,
		"average_distance_meters":1833.33
	}`)
}

// summary decodes the summary endpoint's body.
func summary(t *testing.T, body string) map[string]any {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return decoded
}

func TestSummaryAveragesDistanceOverOnlyTheRidesWithACachedRoute(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 1500.0, 300.0)

	testsupport.Ride(t, db, rides.Ride{Duration: 10, OriginStationCode: "021", DestinationStationCode: "041"})
	testsupport.Ride(t, db, rides.Ride{Duration: 10, OriginStationCode: "021", DestinationStationCode: "999"})

	_, body := get(t, db, "/rides/summary")
	decoded := summary(t, body)

	if decoded["total_rides"] != 2.0 {
		t.Errorf("total_rides = %v, want 2", decoded["total_rides"])
	}

	if decoded["total_distance_meters"] != 1500.0 || decoded["average_distance_meters"] != 1500.0 {
		t.Errorf("distances = %v / %v, want 1500 each", decoded["total_distance_meters"], decoded["average_distance_meters"])
	}
}

func TestSummaryReportsNullDistancesWhenNoRouteIsCachedAtAll(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{Duration: 10})

	_, body := get(t, db, "/rides/summary")
	decoded := summary(t, body)

	if decoded["total_distance_meters"] != nil || decoded["average_distance_meters"] != nil {
		t.Errorf("distances should be null, got %v / %v", decoded["total_distance_meters"], decoded["average_distance_meters"])
	}

	if decoded["total_duration"] != 10.0 {
		t.Errorf("total_duration = %v, want 10", decoded["total_duration"])
	}
}

func TestSummaryRoundsTheAverageDurationToTwoDecimals(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{Duration: 10})
	testsupport.Ride(t, db, rides.Ride{Duration: 11})
	testsupport.Ride(t, db, rides.Ride{Duration: 11})

	_, body := get(t, db, "/rides/summary")

	if average := summary(t, body)["average_duration"]; average != 10.67 {
		t.Errorf("average_duration = %v, want 10.67", average)
	}
}
