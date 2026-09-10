package httpapi

import (
	"log/slog"
	"net/http"

	"velostats/internal/support"
)

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, value any) {
	body, err := support.MarshalJSON(value)
	if err != nil {
		logger.Error("failed to encode response", "error", err)
		writeError(w, http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		logger.Error("failed to write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"message":"` + http.StatusText(status) + `"}`))
}

type results[T any] struct {
	Results []T `json:"results"`
}
