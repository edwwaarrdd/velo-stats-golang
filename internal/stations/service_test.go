package stations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"velostats/internal/stations"
)

// fakeFeed serves a canned GBFS response.
func fakeFeed(t *testing.T, status int, body string) *stations.VeloAntwerpService {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return stations.NewVeloAntwerpService(server.Client(), server.URL)
}

func TestGBFSFeedMapsOntoStations(t *testing.T) {
	service := fakeFeed(t, http.StatusOK, `{"data":{"stations":[{
		"station_id":"021",
		"name":"021- Driekoningen",
		"short_name":"021",
		"lat":51.19548,
		"lon":4.41919,
		"address":"Driekoningenstraat 1",
		"post_code":"2600",
		"rental_methods":["KEY"],
		"capacity":28
	}]}}`)

	fetched, err := service.FetchStations(context.Background())
	if err != nil {
		t.Fatalf("FetchStations: %v", err)
	}

	if len(fetched) != 1 {
		t.Fatalf("fetched %d stations, want 1", len(fetched))
	}

	station := fetched[0]

	if station.StationID != "021" || station.Name != "021- Driekoningen" || station.ShortName != "021" {
		t.Errorf("identity = %+v", station)
	}

	if station.Lat != 51.19548 || station.Lon != 4.41919 {
		t.Errorf("coordinates = %v, %v", station.Lat, station.Lon)
	}

	if station.Address != "Driekoningenstraat 1" || station.PostCode != "2600" || station.Capacity != 28 {
		t.Errorf("details = %+v", station)
	}

	if len(station.RentalMethods) != 1 || station.RentalMethods[0] != "KEY" {
		t.Errorf("rental methods = %v", station.RentalMethods)
	}
}

func TestGBFSFeedDefaultsRentalMethodsAndCapacityWhenOmitted(t *testing.T) {
	service := fakeFeed(t, http.StatusOK, `{"data":{"stations":[{
		"station_id":"021",
		"name":"021- Driekoningen",
		"short_name":"021",
		"lat":51.19548,
		"lon":4.41919,
		"address":"Driekoningenstraat 1",
		"post_code":"2600"
	}]}}`)

	fetched, err := service.FetchStations(context.Background())
	if err != nil {
		t.Fatalf("FetchStations: %v", err)
	}

	if len(fetched[0].RentalMethods) != 0 {
		t.Errorf("rental methods = %v, want empty", fetched[0].RentalMethods)
	}

	if fetched[0].Capacity != 0 {
		t.Errorf("capacity = %d, want 0", fetched[0].Capacity)
	}
}

func TestGBFSFeedFailsOnAnErrorResponse(t *testing.T) {
	service := fakeFeed(t, http.StatusServiceUnavailable, `{}`)

	_, err := service.FetchStations(context.Background())
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Errorf("error = %v, want a failure naming the status", err)
	}
}
