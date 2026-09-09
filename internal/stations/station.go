// Package stations holds the bike-share stations: the model, its storage, the
// upstream feed it is loaded from and the endpoint that lists it.
package stations

// Station is a bike-share station and the place a ride starts or ends.
type Station struct {
	StationID     string
	Name          string
	ShortName     string
	Lat           float64
	Lon           float64
	Address       string
	PostCode      string
	RentalMethods []string
	Capacity      int
}
