package httpapi_test

import (
	"testing"

	"velostats/internal/stations"
	"velostats/internal/testsupport"
)

func TestStationListIsEmptyWhenThereAreNoStations(t *testing.T) {
	_, body := get(t, testsupport.NewDatabase(t), "/stations")

	if body != `{"results":[]}` {
		t.Errorf("body = %s, want an empty result list", body)
	}
}

func TestStationListReturnsEveryStationWithItsCoordinates(t *testing.T) {
	db := testsupport.NewDatabase(t)

	testsupport.Station(t, db, stations.Station{
		StationID: "041",
		Name:      "041- Van Eyck",
		Lat:       51.2189,
		Lon:       4.4131,
	})

	_, body := get(t, db, "/stations")

	assertJSON(t, body, `{"results":[{"station_id":"041","name":"041- Van Eyck","lat":51.2189,"lon":4.4131}]}`)
}
