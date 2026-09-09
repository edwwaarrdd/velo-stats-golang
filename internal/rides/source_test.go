package rides_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"velostats/internal/rides"
)

// writeExport writes a rides export and returns its path.
func writeExport(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rides.json")

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write export: %v", err)
	}

	return path
}

func TestJSONFileSourceMapsTheExportOntoRides(t *testing.T) {
	path := writeExport(t, `{"data":{"CustomerRides":[{
		"id":73147208,
		"accountId":123,
		"status":"Completed",
		"duration":8,
		"bikeNumber":"5097",
		"originStationCode":"021",
		"originStation":"021- Driekoningen",
		"originSlotId":"15",
		"checkoutTime":"2026-09-06 08:57:02",
		"destinationStationCode":"041",
		"destinationStation":"041- Van Eyck",
		"destinationSlotId":"23",
		"checkinTime":"2026-09-06 09:05:30"
	}]}}`)

	fetched, err := rides.NewJSONFileService(path).FetchRides()
	if err != nil {
		t.Fatalf("FetchRides: %v", err)
	}

	if len(fetched) != 1 {
		t.Fatalf("fetched %d rides, want 1", len(fetched))
	}

	ride := fetched[0]

	if ride.RideID != 73147208 || ride.AccountID != 123 || ride.Status != "Completed" || ride.Duration != 8 {
		t.Errorf("ride = %+v", ride)
	}

	if ride.CheckoutTime.Format("2006-01-02 15:04:05") != "2026-09-06 08:57:02" {
		t.Errorf("checkout time = %v", ride.CheckoutTime)
	}

	if ride.CheckoutTime.Location() != ride.CheckinTime.Location() || ride.CheckoutTime.Location().String() != "UTC" {
		t.Errorf("ride times should be in UTC, got %v", ride.CheckoutTime.Location())
	}

	if ride.CheckinTime.Format("2006-01-02 15:04:05") != "2026-09-06 09:05:30" {
		t.Errorf("checkin time = %v", ride.CheckinTime)
	}
}

func TestJSONFileSourceKeepsTheExportOrder(t *testing.T) {
	path := writeExport(t, `{"data":{"CustomerRides":[
		{"id":1,"accountId":1,"status":"Completed","duration":5,"bikeNumber":"1","originStationCode":"021","originStation":"a","originSlotId":"1","checkoutTime":"2026-01-01 08:00:00","destinationStationCode":"041","destinationStation":"b","destinationSlotId":"2","checkinTime":"2026-01-01 08:05:00"},
		{"id":2,"accountId":1,"status":"Completed","duration":5,"bikeNumber":"1","originStationCode":"021","originStation":"a","originSlotId":"1","checkoutTime":"2026-01-02 08:00:00","destinationStationCode":"041","destinationStation":"b","destinationSlotId":"2","checkinTime":"2026-01-02 08:05:00"}
	]}}`)

	fetched, err := rides.NewJSONFileService(path).FetchRides()
	if err != nil {
		t.Fatalf("FetchRides: %v", err)
	}

	if len(fetched) != 2 || fetched[0].RideID != 1 || fetched[1].RideID != 2 {
		t.Errorf("ride ids = %+v", fetched)
	}
}

func TestJSONFileSourceFailsWhenTheExportIsMissing(t *testing.T) {
	_, err := rides.NewJSONFileService("/nowhere/rides.json").FetchRides()

	if err == nil || !strings.Contains(err.Error(), "rides export not found at /nowhere/rides.json") {
		t.Errorf("error = %v, want a missing-export failure", err)
	}
}
