package routing

import (
	"context"

	"velostats/internal/stations"
	"velostats/internal/support"
)

type CachedStationRouteService struct {
	routes  *Repository
	service Service
}

func NewCachedStationRouteService(routes *Repository, service Service) *CachedStationRouteService {
	return &CachedStationRouteService{routes: routes, service: service}
}

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
