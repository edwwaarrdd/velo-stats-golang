package routing

type TravelMode string

const (
	ModeFoot TravelMode = "foot"
	ModeBike TravelMode = "bike"
)

// The demo server at router.project-osrm.org only hosts the car profile and
// silently ignores the profile named in the URL, so every mode came back with
// car driving times. FOSSGIS runs a separate instance per profile, and the
// profile is selected by this path rather than by the URL segment.
func (m TravelMode) OSRMInstancePath() string {
	switch m {
	case ModeFoot:
		return "routed-foot"
	default:
		return "routed-bike"
	}
}
