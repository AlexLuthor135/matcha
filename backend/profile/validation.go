package profile

import (
	"backend/models"
	"strings"
	"time"
)

func isValidGender(gender string) bool {
	switch gender {
	case models.GenderMale, models.GenderFemale, models.GenderOther:
		return true
	default:
		return false
	}
}

func normalizeSexualPreference(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return models.PreferenceEveryone
	}
	return value
}

func isValidSexualPreference(value string) bool {
	switch value {
	case models.PreferenceMale, models.PreferenceFemale, models.PreferenceOther, models.PreferenceEveryone:
		return true
	default:
		return false
	}
}

func isValidAge(birthDate time.Time) bool {
	return !birthDate.After(time.Now().UTC().AddDate(-18, 0, 0))
}

func isLocationValid(latitude *float64, longitude *float64) bool {
	if latitude == nil || longitude == nil {
		return latitude == nil && longitude == nil
	}
	return *latitude >= -90 && *latitude <= 90 && *longitude >= -180 && *longitude <= 180
}
