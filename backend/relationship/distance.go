package relationship

import (
	"math"
	"time"
)

func isLocationValid(latitude *float64, longitude *float64) bool {
	if latitude == nil || longitude == nil {
		return latitude == nil && longitude == nil
	}
	return *latitude >= -90 && *latitude <= 90 && *longitude >= -180 && *longitude <= 180
}

func distanceInKM(firstLatitude float64, firstLongitude float64, secondLatitude float64, secondLongitude float64) float64 {
	const earthRadiusKM = 6371.0
	degreesToRadians := func(value float64) float64 { return value * math.Pi / 180 }
	firstLatitudeRadians := degreesToRadians(firstLatitude)
	secondLatitudeRadians := degreesToRadians(secondLatitude)
	latitudeDifference := degreesToRadians(secondLatitude - firstLatitude)
	longitudeDifference := degreesToRadians(secondLongitude - firstLongitude)
	a := math.Sin(latitudeDifference/2)*math.Sin(latitudeDifference/2) +
		math.Cos(firstLatitudeRadians)*math.Cos(secondLatitudeRadians)*
			math.Sin(longitudeDifference/2)*math.Sin(longitudeDifference/2)
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
