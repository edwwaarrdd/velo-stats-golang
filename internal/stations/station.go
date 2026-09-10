package stations

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
