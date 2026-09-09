// Package routing calculates and caches cycling routes between stations.
package routing

// TravelMode is the way a route is travelled.
type TravelMode string

const (
	ModeFoot TravelMode = "foot"
	ModeBike TravelMode = "bike"
)

// OSRMInstancePath is the OSRM instance that actually routes for this mode.
//
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
