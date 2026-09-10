package stations

import "velostats/internal/support"

type Resource struct {
	StationID string        `json:"station_id"`
	Name      string        `json:"name"`
	Lat       support.Float `json:"lat"`
	Lon       support.Float `json:"lon"`
}

func NewResource(station Station) Resource {
	return Resource{
		StationID: station.StationID,
		Name:      station.Name,
		Lat:       support.Float(station.Lat),
		Lon:       support.Float(station.Lon),
	}
}
