// Package weather fetches and caches the weather a ride was made in.
package weather

import "time"

// Observation is the biking-relevant weather at a place and hour.
type Observation struct {
	TemperatureC            float64
	ApparentTemperatureC    float64
	PrecipitationMm         float64
	RainMm                  float64
	SnowfallCm              float64
	CloudCoverPercent       float64
	WindSpeedKmh            float64
	WindGustsKmh            float64
	WindDirectionDegrees    float64
	RelativeHumidityPercent float64
	WeatherCode             int
	ObservedAt              time.Time
}
