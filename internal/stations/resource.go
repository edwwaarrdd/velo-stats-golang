package stations

import "velostats/internal/support"

// Resource is how a station is rendered by the API.
type Resource struct {
	StationID string        `json:"station_id"`
	Name      string        `json:"name"`
	Lat       support.Float `json:"lat"`
	Lon       support.Float `json:"lon"`
}

// NewResource renders a station.
func NewResource(station Station) Resource {
	return Resource{
		StationID: station.StationID,
		Name:      station.Name,
		Lat:       support.Float(station.Lat),
		Lon:       support.Float(station.Lon),
	}
}
