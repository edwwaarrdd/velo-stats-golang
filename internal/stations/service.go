package stations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// InformationService fetches bike-share station information from upstream.
type InformationService interface {
	FetchStations(ctx context.Context) ([]Station, error)
}

// VeloAntwerpService fetches Velo Antwerp station information from the public
// GBFS feed.
type VeloAntwerpService struct {
	client *http.Client
	url    string
}

// NewVeloAntwerpService builds the GBFS-backed station information service.
func NewVeloAntwerpService(client *http.Client, url string) *VeloAntwerpService {
	return &VeloAntwerpService{client: client, url: url}
}

type gbfsFeed struct {
	Data struct {
		Stations []struct {
			StationID     string   `json:"station_id"`
			Name          string   `json:"name"`
			ShortName     string   `json:"short_name"`
			Lat           float64  `json:"lat"`
			Lon           float64  `json:"lon"`
			Address       string   `json:"address"`
			PostCode      string   `json:"post_code"`
			RentalMethods []string `json:"rental_methods"`
			Capacity      int      `json:"capacity"`
		} `json:"stations"`
	} `json:"data"`
}

// FetchStations returns every station the feed lists.
func (s *VeloAntwerpService) FetchStations(ctx context.Context) ([]Station, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build station information request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch station information: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("station information request failed with status %d", response.StatusCode)
	}

	var feed gbfsFeed
	if err := json.NewDecoder(response.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode station information: %w", err)
	}

	fetched := make([]Station, 0, len(feed.Data.Stations))

	for _, station := range feed.Data.Stations {
		rentalMethods := station.RentalMethods
		if rentalMethods == nil {
			rentalMethods = []string{}
		}

		fetched = append(fetched, Station{
			StationID:     station.StationID,
			Name:          station.Name,
			ShortName:     station.ShortName,
			Lat:           station.Lat,
			Lon:           station.Lon,
			Address:       station.Address,
			PostCode:      station.PostCode,
			RentalMethods: rentalMethods,
			Capacity:      station.Capacity,
		})
	}

	return fetched, nil
}
