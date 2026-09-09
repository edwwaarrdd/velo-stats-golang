package routing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"velostats/internal/routing"
	"velostats/internal/support"
)

var (
	antwerpSouth = support.Coordinate{Lat: 51.19548, Lon: 4.41919}
	antwerpNorth = support.Coordinate{Lat: 51.21797, Lon: 4.40243}
)

// fakeOSRM serves a canned response and records the paths it was asked for.
func fakeOSRM(t *testing.T, body string) (*routing.OSRMService, *[]string) {
	t.Helper()

	requested := &[]string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*requested = append(*requested, r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return routing.NewOSRMService(server.Client(), server.URL), requested
}

func TestOSRMReturnsTheDistanceAndDurationOfTheFirstRoute(t *testing.T) {
	service, _ := fakeOSRM(t, `{"code":"Ok","routes":[{"distance":1502.3,"duration":361.7}]}`)

	route, err := service.GetRoute(context.Background(), antwerpSouth, antwerpNorth, routing.ModeBike)
	if err != nil {
		t.Fatalf("GetRoute: %v", err)
	}

	if route.DistanceMeters != 1502.3 || route.DurationSeconds != 361.7 {
		t.Errorf("route = %+v, want 1502.3m in 361.7s", route)
	}
}

func TestOSRMAsksForTheCoordinatesInLonLatOrder(t *testing.T) {
	service, requested := fakeOSRM(t, `{"code":"Ok","routes":[{"distance":1.0,"duration":1.0}]}`)

	if _, err := service.GetRoute(context.Background(), antwerpSouth, antwerpNorth, routing.ModeBike); err != nil {
		t.Fatalf("GetRoute: %v", err)
	}

	if !strings.Contains((*requested)[0], "/route/v1/bike/4.41919,51.19548;4.40243,51.21797") {
		t.Errorf("requested %s", (*requested)[0])
	}
}

func TestOSRMAsksASeparateInstancePerTravelMode(t *testing.T) {
	service, requested := fakeOSRM(t, `{"code":"Ok","routes":[{"distance":1.0,"duration":1.0}]}`)

	for _, mode := range []routing.TravelMode{routing.ModeBike, routing.ModeFoot} {
		if _, err := service.GetRoute(context.Background(), antwerpSouth, antwerpNorth, mode); err != nil {
			t.Fatalf("GetRoute(%s): %v", mode, err)
		}
	}

	if !strings.Contains((*requested)[0], "/routed-bike/") {
		t.Errorf("the bike request went to %s", (*requested)[0])
	}

	if !strings.Contains((*requested)[1], "/routed-foot/") {
		t.Errorf("the foot request went to %s", (*requested)[1])
	}
}

func TestOSRMFailsWhenItCannotRouteBetweenTheCoordinates(t *testing.T) {
	service, _ := fakeOSRM(t, `{"code":"NoRoute","message":"no route found"}`)

	_, err := service.GetRoute(context.Background(), antwerpSouth, antwerpNorth, routing.ModeBike)
	if err == nil || !strings.Contains(err.Error(), "OSRM request failed: no route found") {
		t.Errorf("error = %v, want an OSRM failure naming the message", err)
	}
}
