package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"velostats/internal/rides"
	"velostats/internal/routing"
	"velostats/internal/testsupport"
	"velostats/internal/weather"
)

func TestRideListIsEmptyWhenThereAreNoRides(t *testing.T) {
	status, body := get(t, testsupport.NewDatabase(t), "/rides")

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	if body != `{"results":[]}` {
		t.Errorf("body = %s, want an empty result list", body)
	}
}

func TestRideListReturnsRidesMostRecentFirst(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{RideID: 1, CheckoutTime: testsupport.Time(t, "2026-01-01 08:00:00")})
	testsupport.Ride(t, db, rides.Ride{RideID: 2, CheckoutTime: testsupport.Time(t, "2026-03-01 08:00:00")})
	testsupport.Ride(t, db, rides.Ride{RideID: 3, CheckoutTime: testsupport.Time(t, "2026-02-01 08:00:00")})

	_, body := get(t, db, "/rides")

	var payload struct {
		Results []struct {
			RideID int64 `json:"ride_id"`
		} `json:"results"`
	}

	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	got := []int64{}
	for _, result := range payload.Results {
		got = append(got, result.RideID)
	}

	if len(got) != 3 || got[0] != 2 || got[1] != 3 || got[2] != 1 {
		t.Errorf("ride ids = %v, want [2 3 1]", got)
	}
}

func TestRideListReturnsEveryFieldWithDistanceSpeedExpectedTimeAndWeather(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 1500.0, 400.0)

	ride := testsupport.Ride(t, db, rides.Ride{
		RideID:                 73147208,
		AccountID:              123,
		Status:                 "Completed",
		Duration:               10,
		BikeNumber:             "5097",
		OriginStationCode:      "021",
		OriginStation:          "021- Driekoningen",
		OriginSlotID:           "15",
		CheckoutTime:           testsupport.Time(t, "2026-09-06 08:57:02"),
		DestinationStationCode: "041",
		DestinationStation:     "041- Van Eyck",
		DestinationSlotID:      "23",
		CheckinTime:            testsupport.Time(t, "2026-09-06 09:05:30"),
	})

	testsupport.WeatherRecord(t, db, ride.RideID, weather.Observation{
		TemperatureC:            18.0,
		ApparentTemperatureC:    17.1,
		PrecipitationMm:         0.0,
		RainMm:                  0.0,
		SnowfallCm:              0.0,
		CloudCoverPercent:       42.0,
		WindSpeedKmh:            11.2,
		WindGustsKmh:            24.5,
		WindDirectionDegrees:    210.0,
		RelativeHumidityPercent: 68.0,
		WeatherCode:             3,
		ObservedAt:              testsupport.Time(t, "2026-09-06 09:00:00"),
	})

	status, body := get(t, db, "/rides")

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	assertJSON(t, body, `{"results":[{
		"ride_id":73147208,
		"account_id":123,
		"status":"Completed",
		"duration":10,
		"bike_number":"5097",
		"origin_station_code":"021",
		"origin_station":"021- Driekoningen",
		"origin_slot_id":"15",
		"checkout_time":"2026-09-06T08:57:02Z",
		"destination_station_code":"041",
		"destination_station":"041- Van Eyck",
		"destination_slot_id":"23",
		"checkin_time":"2026-09-06T09:05:30Z",
		"distance_meters":1500.0,
		"speed_kmh":10.63,
		"expected_duration_seconds":400.0,
		"actual_duration_seconds":508.0,
		"duration_vs_expected_seconds":108.0,
		"weather":{
			"temperature_c":18.0,
			"apparent_temperature_c":17.1,
			"precipitation_mm":0.0,
			"rain_mm":0.0,
			"snowfall_cm":0.0,
			"cloud_cover_percent":42.0,
			"wind_speed_kmh":11.2,
			"wind_gusts_kmh":24.5,
			"wind_direction_degrees":210.0,
			"relative_humidity_percent":68.0,
			"weather_code":3,
			"observed_at":"2026-09-06T09:00:00Z"
		}
	}]}`)
}

func firstResult(t *testing.T, body string) map[string]any {
	t.Helper()

	var payload struct {
		Results []map[string]any `json:"results"`
	}

	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Results) == 0 {
		t.Fatalf("the listing returned no results")
	}

	return payload.Results[0]
}

func TestRideListReturnsNullDistanceAndSpeedWhenNoRouteIsCached(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Ride(t, db, rides.Ride{OriginStationCode: "021", DestinationStationCode: "041"})

	_, body := get(t, db, "/rides")
	result := firstResult(t, body)

	for _, field := range []string{"distance_meters", "speed_kmh", "expected_duration_seconds", "duration_vs_expected_seconds", "weather"} {
		if result[field] != nil {
			t.Errorf("%s = %v, want null", field, result[field])
		}
	}
}

func TestRideListReportsANegativeDeltaWhenTheRideBeatTheExpectedTime(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 1500.0, 400.0)
	testsupport.Ride(t, db, rides.Ride{
		OriginStationCode:      "021",
		DestinationStationCode: "041",
		CheckoutTime:           testsupport.Time(t, "2026-09-06 08:57:00"),
		CheckinTime:            testsupport.Time(t, "2026-09-06 09:02:00"),
	})

	_, body := get(t, db, "/rides")
	result := firstResult(t, body)

	// 300 seconds ridden against the 400 seconds the router predicted.
	if result["actual_duration_seconds"] != 300.0 {
		t.Errorf("actual_duration_seconds = %v, want 300", result["actual_duration_seconds"])
	}

	if result["duration_vs_expected_seconds"] != -100.0 {
		t.Errorf("duration_vs_expected_seconds = %v, want -100", result["duration_vs_expected_seconds"])
	}
}

func TestRideListIgnoresRoutesCachedForAnotherTravelMode(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.CachedRoute(t, db, "021", "041", routing.ModeFoot, 1500.0, 300.0)
	testsupport.Ride(t, db, rides.Ride{OriginStationCode: "021", DestinationStationCode: "041"})

	_, body := get(t, db, "/rides")

	if distance := firstResult(t, body)["distance_meters"]; distance != nil {
		t.Errorf("distance_meters = %v, want null", distance)
	}
}

func TestRideListReturnsNullSpeedWhenNoTimePassed(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 1500.0, 300.0)
	testsupport.Ride(t, db, rides.Ride{
		OriginStationCode:      "021",
		DestinationStationCode: "041",
		Duration:               1,
		CheckoutTime:           testsupport.Time(t, "2026-09-06 08:57:00"),
		CheckinTime:            testsupport.Time(t, "2026-09-06 08:57:00"),
	})

	_, body := get(t, db, "/rides")
	result := firstResult(t, body)

	if result["distance_meters"] != 1500.0 {
		t.Errorf("distance_meters = %v, want 1500", result["distance_meters"])
	}

	if result["speed_kmh"] != nil {
		t.Errorf("speed_kmh = %v, want null", result["speed_kmh"])
	}
}

func TestRideListRoundsTheSpeedToTwoDecimals(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 2345.0, 300.0)
	testsupport.Ride(t, db, rides.Ride{
		OriginStationCode:      "021",
		DestinationStationCode: "041",
		CheckoutTime:           testsupport.Time(t, "2026-09-06 08:00:00"),
		CheckinTime:            testsupport.Time(t, "2026-09-06 08:06:59"),
	})

	_, body := get(t, db, "/rides")

	// 2.345 km in 419 seconds is 20.14_ km/h.
	if speed := firstResult(t, body)["speed_kmh"]; speed != 20.15 {
		t.Errorf("speed_kmh = %v, want 20.15", speed)
	}
}

func TestRideListBasesTheSpeedOnExactSecondsRatherThanTheRoundedDuration(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.BikeRouteBetween(t, db, "021", "041", 1742.4, 248.1)
	testsupport.Ride(t, db, rides.Ride{
		OriginStationCode:      "021",
		DestinationStationCode: "041",
		// The stored duration truncates 4m29s to 4 whole minutes, which would
		// overstate the speed as 26.14 km/h.
		Duration:     4,
		CheckoutTime: testsupport.Time(t, "2026-09-06 08:00:00"),
		CheckinTime:  testsupport.Time(t, "2026-09-06 08:04:29"),
	})

	_, body := get(t, db, "/rides")

	if speed := firstResult(t, body)["speed_kmh"]; speed != 23.32 {
		t.Errorf("speed_kmh = %v, want 23.32", speed)
	}
}
