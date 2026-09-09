package weather

import "velostats/internal/support"

// Resource is how an observation is rendered by the API.
type Resource struct {
	TemperatureC            support.Float `json:"temperature_c"`
	ApparentTemperatureC    support.Float `json:"apparent_temperature_c"`
	PrecipitationMm         support.Float `json:"precipitation_mm"`
	RainMm                  support.Float `json:"rain_mm"`
	SnowfallCm              support.Float `json:"snowfall_cm"`
	CloudCoverPercent       support.Float `json:"cloud_cover_percent"`
	WindSpeedKmh            support.Float `json:"wind_speed_kmh"`
	WindGustsKmh            support.Float `json:"wind_gusts_kmh"`
	WindDirectionDegrees    support.Float `json:"wind_direction_degrees"`
	RelativeHumidityPercent support.Float `json:"relative_humidity_percent"`
	WeatherCode             int           `json:"weather_code"`
	ObservedAt              string        `json:"observed_at"`
}

// NewResource renders an observation.
func NewResource(observation Observation) Resource {
	return Resource{
		TemperatureC:            support.Float(observation.TemperatureC),
		ApparentTemperatureC:    support.Float(observation.ApparentTemperatureC),
		PrecipitationMm:         support.Float(observation.PrecipitationMm),
		RainMm:                  support.Float(observation.RainMm),
		SnowfallCm:              support.Float(observation.SnowfallCm),
		CloudCoverPercent:       support.Float(observation.CloudCoverPercent),
		WindSpeedKmh:            support.Float(observation.WindSpeedKmh),
		WindGustsKmh:            support.Float(observation.WindGustsKmh),
		WindDirectionDegrees:    support.Float(observation.WindDirectionDegrees),
		RelativeHumidityPercent: support.Float(observation.RelativeHumidityPercent),
		WeatherCode:             observation.WeatherCode,
		ObservedAt:              support.APIDateTime(observation.ObservedAt),
	}
}
