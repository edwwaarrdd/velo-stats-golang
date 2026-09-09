package weather

import (
	"context"
	"time"

	"velostats/internal/support"
)

// CachedRideWeatherService fetches the weather for a ride's checkin time and
// origin station, caching results so the same ride's weather is only ever
// fetched once unless forced.
type CachedRideWeatherService struct {
	records *Repository
	service Service
}

// NewCachedRideWeatherService builds the caching ride weather service.
func NewCachedRideWeatherService(records *Repository, service Service) *CachedRideWeatherService {
	return &CachedRideWeatherService{records: records, service: service}
}

// GetWeather returns the cached observation for a ride, fetching and caching it
// when it is not known yet or when the fetch is forced.
func (s *CachedRideWeatherService) GetWeather(
	ctx context.Context,
	rideID int64,
	checkinTime time.Time,
	location support.Coordinate,
	force bool,
) (Observation, error) {
	if !force {
		cached, err := s.records.FindForRide(ctx, rideID)
		if err != nil {
			return Observation{}, err
		}

		if cached != nil {
			return *cached, nil
		}
	}

	observation, err := s.service.GetWeather(ctx, location, checkinTime)
	if err != nil {
		return Observation{}, err
	}

	if err := s.records.Save(ctx, rideID, observation); err != nil {
		return Observation{}, err
	}

	return observation, nil
}
