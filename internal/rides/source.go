package rides

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type DataSource interface {
	FetchRides() ([]Ride, error)
}

type JSONFileService struct {
	path string
}

func NewJSONFileService(path string) *JSONFileService {
	return &JSONFileService{path: path}
}

const rideDateTimeFormat = "2006-01-02 15:04:05"

type ridesExport struct {
	Data struct {
		CustomerRides []struct {
			ID                     int64  `json:"id"`
			AccountID              int64  `json:"accountId"`
			Status                 string `json:"status"`
			Duration               int    `json:"duration"`
			BikeNumber             string `json:"bikeNumber"`
			OriginStationCode      string `json:"originStationCode"`
			OriginStation          string `json:"originStation"`
			OriginSlotID           string `json:"originSlotId"`
			CheckoutTime           string `json:"checkoutTime"`
			DestinationStationCode string `json:"destinationStationCode"`
			DestinationStation     string `json:"destinationStation"`
			DestinationSlotID      string `json:"destinationSlotId"`
			CheckinTime            string `json:"checkinTime"`
		} `json:"CustomerRides"`
	} `json:"data"`
}

func (s *JSONFileService) FetchRides() ([]Ride, error) {
	payload, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("rides export not found at %s", s.path)
	}

	if err != nil {
		return nil, fmt.Errorf("read rides export: %w", err)
	}

	var export ridesExport
	if err := json.Unmarshal(payload, &export); err != nil {
		return nil, fmt.Errorf("decode rides export: %w", err)
	}

	fetched := make([]Ride, 0, len(export.Data.CustomerRides))

	for _, ride := range export.Data.CustomerRides {
		checkoutTime, err := parseExportTime(ride.CheckoutTime)
		if err != nil {
			return nil, err
		}

		checkinTime, err := parseExportTime(ride.CheckinTime)
		if err != nil {
			return nil, err
		}

		fetched = append(fetched, Ride{
			RideID:                 ride.ID,
			AccountID:              ride.AccountID,
			Status:                 ride.Status,
			Duration:               ride.Duration,
			BikeNumber:             ride.BikeNumber,
			OriginStationCode:      ride.OriginStationCode,
			OriginStation:          ride.OriginStation,
			OriginSlotID:           ride.OriginSlotID,
			CheckoutTime:           checkoutTime,
			DestinationStationCode: ride.DestinationStationCode,
			DestinationStation:     ride.DestinationStation,
			DestinationSlotID:      ride.DestinationSlotID,
			CheckinTime:            checkinTime,
		})
	}

	return fetched, nil
}

func parseExportTime(value string) (time.Time, error) {
	parsed, err := time.ParseInLocation(rideDateTimeFormat, value, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse ride time %q: %w", value, err)
	}

	return parsed, nil
}
