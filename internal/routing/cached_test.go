package routing_test

import (
	"context"
	"errors"
	"testing"

	"velostats/internal/routing"
	"velostats/internal/stations"
	"velostats/internal/support"
	"velostats/internal/testsupport"
)

type countingService struct {
	route routing.Route
	calls int
	err   error
}

func (s *countingService) GetRoute(context.Context, support.Coordinate, support.Coordinate, routing.TravelMode) (routing.Route, error) {
	s.calls++

	return s.route, s.err
}

func TestCachedRouteServiceOnlyCalculatesARoutePairOnce(t *testing.T) {
	db := testsupport.NewDatabase(t)

	origin := testsupport.Station(t, db, stations.Station{StationID: "021"})
	destination := testsupport.Station(t, db, stations.Station{StationID: "041"})

	upstream := &countingService{route: routing.Route{DistanceMeters: 1502.3, DurationSeconds: 361.7}}
	service := routing.NewCachedStationRouteService(routing.NewRepository(db), upstream)

	for range 3 {
		route, err := service.GetRoute(context.Background(), origin, destination, routing.ModeBike)
		if err != nil {
			t.Fatalf("GetRoute: %v", err)
		}

		if route.DistanceMeters != 1502.3 || route.DurationSeconds != 361.7 {
			t.Errorf("route = %+v", route)
		}
	}

	if upstream.calls != 1 {
		t.Errorf("the router was called %d times, want 1", upstream.calls)
	}
}

func TestCachedRouteServiceCachesEachTravelModeSeparately(t *testing.T) {
	db := testsupport.NewDatabase(t)

	origin := testsupport.Station(t, db, stations.Station{StationID: "021"})
	destination := testsupport.Station(t, db, stations.Station{StationID: "041"})

	upstream := &countingService{route: routing.Route{DistanceMeters: 1.0, DurationSeconds: 1.0}}
	service := routing.NewCachedStationRouteService(routing.NewRepository(db), upstream)

	for _, mode := range []routing.TravelMode{routing.ModeBike, routing.ModeFoot} {
		if _, err := service.GetRoute(context.Background(), origin, destination, mode); err != nil {
			t.Fatalf("GetRoute(%s): %v", mode, err)
		}
	}

	if upstream.calls != 2 {
		t.Errorf("the router was called %d times, want one call per mode", upstream.calls)
	}
}

func TestCachedRouteServiceCachesNothingWhenTheRouterFails(t *testing.T) {
	db := testsupport.NewDatabase(t)

	origin := testsupport.Station(t, db, stations.Station{StationID: "021"})
	destination := testsupport.Station(t, db, stations.Station{StationID: "041"})

	upstream := &countingService{err: errors.New("upstream is down")}
	service := routing.NewCachedStationRouteService(routing.NewRepository(db), upstream)

	if _, err := service.GetRoute(context.Background(), origin, destination, routing.ModeBike); err == nil {
		t.Fatal("GetRoute should have failed")
	}

	cached, err := routing.NewRepository(db).Find(context.Background(), "021", "041", routing.ModeBike)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	if cached != nil {
		t.Errorf("a failed lookup should not be cached, got %+v", cached)
	}
}
