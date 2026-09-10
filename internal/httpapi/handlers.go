package httpapi

import (
	"log/slog"
	"net/http"

	"velostats/internal/rides"
	"velostats/internal/stations"
)

// RideHandler serves the ride endpoints.
type RideHandler struct {
	rides   *rides.Repository
	summary *rides.SummaryCalculator
	cost    *rides.CostCalculator
	logger  *slog.Logger
}

// NewRideHandler builds the ride endpoints.
func NewRideHandler(repository *rides.Repository, summary *rides.SummaryCalculator, cost *rides.CostCalculator, logger *slog.Logger) *RideHandler {
	return &RideHandler{rides: repository, summary: summary, cost: cost, logger: logger}
}

// Index lists every ride, most recent first, with its cached distance, expected
// ride time and weather.
func (h *RideHandler) Index(w http.ResponseWriter, r *http.Request) {
	listed, err := h.rides.All(r.Context())
	if err != nil {
		h.logger.Error("failed to list rides", "error", err)
		writeError(w, http.StatusInternalServerError)

		return
	}

	resources := make([]rides.Resource, 0, len(listed))
	for _, ride := range listed {
		resources = append(resources, rides.NewResource(ride))
	}

	writeJSON(w, h.logger, http.StatusOK, results[rides.Resource]{Results: resources})
}

// Summary reports aggregate duration and distance statistics across every ride.
func (h *RideHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.summary.Calculate(r.Context())
	if err != nil {
		h.logger.Error("failed to summarise rides", "error", err)
		writeError(w, http.StatusInternalServerError)

		return
	}

	writeJSON(w, h.logger, http.StatusOK, summary)
}

// Cost reports the subscription cost per ride, and how it compares to passes.
func (h *RideHandler) Cost(w http.ResponseWriter, r *http.Request) {
	cost, err := h.cost.Calculate(r.Context())
	if err != nil {
		h.logger.Error("failed to calculate ride cost", "error", err)
		writeError(w, http.StatusInternalServerError)

		return
	}

	writeJSON(w, h.logger, http.StatusOK, cost)
}

// StationHandler serves the station endpoints.
type StationHandler struct {
	stations *stations.Repository
	logger   *slog.Logger
}

// NewStationHandler builds the station endpoints.
func NewStationHandler(repository *stations.Repository, logger *slog.Logger) *StationHandler {
	return &StationHandler{stations: repository, logger: logger}
}

// Index lists every known station with its coordinates.
func (h *StationHandler) Index(w http.ResponseWriter, r *http.Request) {
	found, err := h.stations.All(r.Context())
	if err != nil {
		h.logger.Error("failed to list stations", "error", err)
		writeError(w, http.StatusInternalServerError)

		return
	}

	resources := make([]stations.Resource, 0, len(found))
	for _, station := range found {
		resources = append(resources, stations.NewResource(station))
	}

	writeJSON(w, h.logger, http.StatusOK, results[stations.Resource]{Results: resources})
}

// healthcheck reports that the app is up.
func healthcheck(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, logger, http.StatusOK, map[string]string{"message": "ok"})
	}
}
