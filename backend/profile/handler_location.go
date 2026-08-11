package profile

import "backend/models"

type LocationRequest struct {
	Source    string   `json:"source"`
	Name      string   `json:"name"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Consent   bool     `json:"consent"`
}

func locationInput(request *LocationRequest) *LocationInput {
	if request == nil {
		return nil
	}
	return &LocationInput{
		Source:    models.LocationSource(request.Source),
		Name:      request.Name,
		Latitude:  request.Latitude,
		Longitude: request.Longitude,
		Consent:   request.Consent,
	}
}
