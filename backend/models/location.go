package models

type LocationSource string

const (
	LocationSourceGPS    LocationSource = "gps"
	LocationSourceManual LocationSource = "manual"
)

func (source LocationSource) IsValid() bool {
	switch source {
	case LocationSourceGPS, LocationSourceManual:
		return true
	default:
		return false
	}
}
