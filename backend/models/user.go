package models

import (
	"time"
)

const (
	GenderMale   = "Male"
	GenderFemale = "Female"
	GenderOther  = "Other"

	PreferenceMale     = "Male"
	PreferenceFemale   = "Female"
	PreferenceOther    = "Other"
	PreferenceEveryone = "Everyone"
)

type User struct {
	ID                uint           `json:"id"`
	IsVerified        bool           `json:"is_verified"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	LastSeenAt        time.Time      `json:"last_seen_at"`
	BirthDate         *time.Time     `json:"birth_date"`
	Latitude          *float64       `json:"latitude"`
	Longitude         *float64       `json:"longitude"`
	LocationSource    LocationSource `json:"location_source"`
	LocationName      string         `json:"location_name"`
	LocationConsentAt *time.Time     `json:"location_consent_at"`
	UserName          string         `json:"user_name"`
	FirstName         string         `json:"first_name"`
	LastName          string         `json:"last_name"`
	Email             string         `json:"email"`
	Password          string         `json:"-"`
	IsCompleted       bool           `json:"is_completed"`
	Gender            string         `json:"gender"`
	Preferences       string         `json:"preferences"`
	Bio               string         `json:"bio"`
	Interests         []string       `json:"interests"`
	Avatar            string         `json:"avatar"`
	Photos            []Photo        `json:"photos,omitempty"`
	FameRating        int64          `json:"fame_rating"`
	Distance          *float64       `json:"distance,omitempty"`
}
