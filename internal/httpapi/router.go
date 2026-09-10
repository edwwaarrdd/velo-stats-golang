package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"velostats/internal/rides"
	"velostats/internal/stations"
)

func Router(
	rideRepository *rides.Repository,
	summary *rides.SummaryCalculator,
	cost *rides.CostCalculator,
	stationRepository *stations.Repository,
	allowedOrigins []string,
	logger *slog.Logger,
) http.Handler {
	rideHandler := NewRideHandler(rideRepository, summary, cost, logger)
	stationHandler := NewStationHandler(stationRepository, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_healthcheck", healthcheck(logger))
	mux.HandleFunc("GET /rides", rideHandler.Index)
	mux.HandleFunc("GET /rides/summary", rideHandler.Summary)
	mux.HandleFunc("GET /rides/cost", rideHandler.Cost)
	mux.HandleFunc("GET /stations", stationHandler.Index)

	return normalisePath(cors(allowedOrigins, logRequests(logger, mux)))
}

func normalisePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalised := "/" + strings.Trim(r.URL.Path, "/")

		if normalised != r.URL.Path {
			r = r.Clone(r.Context())
			r.URL.Path = normalised
			// The escaped form is now stale, so let it be derived from the path.
			r.URL.RawPath = ""
		}

		next.ServeHTTP(w, r)
	})
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
