package routing

import (
	"context"

	"velostats/internal/stations"
	"velostats/internal/support"
)

// CachedStationRouteService calculates routes between stations, caching results
// so a route between the same pair of stations and travel mode is only ever
// calculated once.
type CachedStationRouteService struct {
	routes  *Repository
	service Service
}

// NewCachedStationRouteService builds the caching route service.
func NewCachedStationRouteService(routes *Repository, service Service) *CachedStationRouteService {
	return &CachedStationRouteService{routes: routes, service: service}
}

// GetRoute returns the cached route between two stations, calculating and
// caching it when it is not known yet.
func (s *CachedStationRouteService) GetRoute(ctx context.Context, origin, destination stations.Station, mode TravelMode) (Route, error) {
	cached, err := s.routes.Find(ctx, origin.StationID, destination.StationID, mode)
	if err != nil {
		return Route{}, err
	}

	if cached != nil {
		return *cached, nil
	}

	route, err := s.service.GetRoute(
		ctx,
		support.Coordinate{Lat: origin.Lat, Lon: origin.Lon},
		support.Coordinate{Lat: destination.Lat, Lon: destination.Lon},
		mode,
	)
	if err != nil {
		return Route{}, err
	}

	if err := s.routes.Save(ctx, origin.StationID, destination.StationID, mode, route); err != nil {
		return Route{}, err
	}

	return route, nil
}
