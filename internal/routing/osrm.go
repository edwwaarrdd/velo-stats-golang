package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"velostats/internal/support"
)

type Service interface {
	GetRoute(ctx context.Context, origin, destination support.Coordinate, mode TravelMode) (Route, error)
}

type OSRMService struct {
	client  *http.Client
	baseURL string
}

func NewOSRMService(client *http.Client, baseURL string) *OSRMService {
	return &OSRMService{client: client, baseURL: baseURL}
}

type osrmResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Routes  []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

func (s *OSRMService) GetRoute(ctx context.Context, origin, destination support.Coordinate, mode TravelMode) (Route, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// OSRM expects coordinates as "lon,lat", not "lat,lon".
	coordinates := fmt.Sprintf("%v,%v;%v,%v", origin.Lon, origin.Lat, destination.Lon, destination.Lat)

	url := fmt.Sprintf("%s/%s/route/v1/%s/%s?overview=false", s.baseURL, mode.OSRMInstancePath(), mode, coordinates)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Route{}, fmt.Errorf("build OSRM request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return Route{}, fmt.Errorf("OSRM request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Route{}, fmt.Errorf("OSRM request failed with status %d", response.StatusCode)
	}

	var payload osrmResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Route{}, fmt.Errorf("decode OSRM response: %w", err)
	}

	if payload.Code != "Ok" {
		return Route{}, fmt.Errorf("OSRM request failed: %s", firstNonEmpty(payload.Message, payload.Code, "unknown error"))
	}

	if len(payload.Routes) == 0 {
		return Route{}, fmt.Errorf("OSRM request failed: no routes returned")
	}

	return Route{
		DistanceMeters:  payload.Routes[0].Distance,
		DurationSeconds: payload.Routes[0].Duration,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
