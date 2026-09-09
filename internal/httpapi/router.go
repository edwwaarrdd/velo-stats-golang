package httpapi

import (
	"log/slog"
	"net/http"

	"velostats/internal/rides"
	"velostats/internal/stations"
)

// Router builds the API's routes.
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

	return cors(allowedOrigins, logRequests(logger, mux))
}

// logRequests writes one line per request.
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
