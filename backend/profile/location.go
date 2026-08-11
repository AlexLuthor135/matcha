package profile

import (
	"backend/models"
	"strings"
	"time"
)

type LocationInput struct {
	Source    models.LocationSource
	Name      string
	Latitude  *float64
	Longitude *float64
	Consent   bool
	ConsentAt *time.Time
}

func (input *LocationInput) Prepare() error {
	input.Source = models.LocationSource(strings.ToLower(strings.TrimSpace(string(input.Source))))
	input.Name = strings.TrimSpace(input.Name)
	if !input.Source.IsValid() {
		return ProfileErrors.InvalidLocationSource
	}
	if input.Latitude == nil || input.Longitude == nil || !isLocationValid(input.Latitude, input.Longitude) {
		return ProfileErrors.InvalidLocation
	}
	switch input.Source {
	case models.LocationSourceGPS:
		if !input.Consent {
			return ProfileErrors.LocationConsentRequired
		}
		consentAt := time.Now().UTC()
		input.ConsentAt = &consentAt
		input.Name = ""
	case models.LocationSourceManual:
		if input.Name == "" {
			return ProfileErrors.ManualLocationNameMissing
		}
		input.Consent = false
		input.ConsentAt = nil
	}
	return nil
}
