package weather

import (
	"context"
	"time"

	"velostats/internal/support"
)

type CachedRideWeatherService struct {
	records *Repository
	service Service
}

func NewCachedRideWeatherService(records *Repository, service Service) *CachedRideWeatherService {
	return &CachedRideWeatherService{records: records, service: service}
}

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
