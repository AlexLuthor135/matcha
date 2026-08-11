package user

import (
	"backend/models"
	"math"
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
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
	adultBoundary := time.Now().UTC().AddDate(-18, 0, 0)
	if birthDate.After(adultBoundary) {
		return false
	}
	return true
}

func isValidUserName(userName string) bool {
	length := utf8.RuneCountInString(userName)
	if length < 3 || length > 30 {
		return false
	}
	for _, character := range userName {
		if unicode.IsLetter(character) ||
			unicode.IsDigit(character) ||
			character == '_' ||
			character == '-' ||
			character == '.' {
			continue
		}
		return false
	}
	return true
}

func isValidPassword(password string) bool {
	if len(password) < 8 || len([]byte(password)) > 72 {
		return false
	}
	if isCommonPassword(password) {
		return false
	}
	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, character := range password {
		switch {
		case unicode.IsSpace(character):
			return false
		case unicode.IsUpper(character):
			hasUpper = true
		case unicode.IsLower(character):
			hasLower = true
		case unicode.IsDigit(character):
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*", character):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}

func isLocationValid(latitude *float64, longitude *float64) bool {
	if latitude == nil || longitude == nil {
		return latitude == nil && longitude == nil
	}
	return *latitude >= -90 && *latitude <= 90 && *longitude >= -180 && *longitude <= 180
}

func distanceInKM(WP1_latitude float64, WP1_longitude float64, WP2_latitude float64, WP2_longitude float64) float64 {
	const earthRadiusKM = 6371.0
	degreesToRadians := func(value float64) float64 {
		return value * math.Pi / 180
	}
	WP1_LatitudeRadians := degreesToRadians(WP1_latitude)
	WP2_LatitudeRadians := degreesToRadians(WP2_latitude)

	latitudeDifference := degreesToRadians(WP2_latitude - WP1_latitude)
	longitudeDifference := degreesToRadians(WP2_longitude - WP1_longitude)

	a := math.Sin(latitudeDifference/2)*
		math.Sin(latitudeDifference/2) +
		math.Cos(WP1_LatitudeRadians)*
			math.Cos(WP2_LatitudeRadians)*
			math.Sin(longitudeDifference/2)*
			math.Sin(longitudeDifference/2)
	a = math.Max(0, math.Min(1, a))
	distance := earthRadiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return math.Round(distance*10) / 10
}

func ageAt(birthDate time.Time) int {
	birthDate = birthDate.UTC()
	currentDate := time.Now().UTC()
	age := currentDate.Year() - birthDate.Year()
	birthdayThisYear := time.Date(currentDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
	if currentDate.Before(birthdayThisYear) {
		age--
	}
	return age
}

func isValidEmail(email string) bool {
	if len(email) > 254 || strings.ContainsAny(email, "\r\n") {
		return false
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}
	atIndex := strings.LastIndexByte(email, '@')
	if atIndex <= 0 || atIndex == len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]
	return strings.Contains(domain, ".")
}
