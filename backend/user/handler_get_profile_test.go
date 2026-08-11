package user

import (
	"backend/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProfileBirthDateResponse(t *testing.T) {
	if got := profileBirthDateResponse(nil); got != nil {
		t.Fatalf("profileBirthDateResponse(nil) = %v, want nil", got)
	}

	birthDate := time.Date(2000, time.May, 10, 0, 0, 0, 0, time.UTC)
	got := profileBirthDateResponse(&birthDate)
	if got == nil || *got != "2000-05-10" {
		t.Fatalf("profileBirthDateResponse() = %v, want 2000-05-10", got)
	}
}

func TestProfileAge(t *testing.T) {
	if got := profileAge(nil); got != nil {
		t.Fatalf("profileAge(nil) = %v, want nil", got)
	}

	now := time.Now().UTC()
	tests := []struct {
		name      string
		birthDate time.Time
		wantAge   int
	}{
		{
			name:      "birthday today",
			birthDate: time.Date(now.Year()-30, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			wantAge:   30,
		},
		{
			name:      "birthday already passed",
			birthDate: now.AddDate(-30, 0, -1),
			wantAge:   30,
		},
		{
			name:      "birthday not reached",
			birthDate: now.AddDate(-30, 0, 1),
			wantAge:   29,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := profileAge(&test.birthDate)
			if got == nil || *got != test.wantAge {
				t.Fatalf("profileAge(%v) = %v, want %d", test.birthDate, got, test.wantAge)
			}
		})
	}
}

func TestGetProfileReturnsPrivateBirthDateAndLocation(t *testing.T) {
	const userID uint = 10
	birthDate := time.Date(2000, time.May, 10, 0, 0, 0, 0, time.UTC)
	latitude := 52.52
	longitude := 13.405
	consentAt := time.Now().UTC()
	repository := &fakeUserRepository{
		getProfileFn: func(_ context.Context, gotUserID uint) (models.User, error) {
			if gotUserID != userID {
				t.Fatalf("GetProfile() userID = %d, want %d", gotUserID, userID)
			}
			return models.User{
				ID:                userID,
				UserName:          "private-user",
				BirthDate:         &birthDate,
				Latitude:          &latitude,
				Longitude:         &longitude,
				LocationSource:    models.LocationSourceGPS,
				LocationConsentAt: &consentAt,
				Interests:         []string{},
				Photos:            []models.Photo{},
				FameRating:        7,
			}, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile", nil),
		userID,
	)
	response := httptest.NewRecorder()

	handler.GetProfile(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusOK, response.Body.String())
	}
	var body ProfileResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.BirthDate == nil || *body.BirthDate != "2000-05-10" {
		t.Fatalf("birth_date = %v, want 2000-05-10", body.BirthDate)
	}
	if body.Latitude == nil || *body.Latitude != latitude {
		t.Fatalf("latitude = %v, want %v", body.Latitude, latitude)
	}
	if body.Longitude == nil || *body.Longitude != longitude {
		t.Fatalf("longitude = %v, want %v", body.Longitude, longitude)
	}
	if body.LocationSource != models.LocationSourceGPS {
		t.Fatalf("location_source = %q, want %q", body.LocationSource, models.LocationSourceGPS)
	}
	if body.LocationName != "" {
		t.Fatalf("location_name = %q, want empty GPS location name", body.LocationName)
	}
	if body.LocationConsentAt == nil || !body.LocationConsentAt.Equal(consentAt) {
		t.Fatalf("location_consent_at = %v, want %v", body.LocationConsentAt, consentAt)
	}
}

func TestPublicProfileResponseDoesNotExposeBirthDateOrCoordinates(t *testing.T) {
	birthDate := time.Date(2000, time.May, 10, 0, 0, 0, 0, time.UTC)
	latitude := 52.52
	longitude := 13.405
	data, err := json.Marshal(publicProfileResponse(models.User{
		BirthDate: &birthDate,
		Latitude:  &latitude,
		Longitude: &longitude,
	}))
	if err != nil {
		t.Fatalf("marshal public profile response: %v", err)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode public profile response: %v", err)
	}
	for _, privateField := range []string{
		"birth_date",
		"latitude",
		"longitude",
		"location_source",
		"location_name",
		"location_consent_at",
	} {
		if _, exists := body[privateField]; exists {
			t.Fatalf("public profile exposes private field %q: %s", privateField, data)
		}
	}
	if _, exists := body["age"]; !exists {
		t.Fatalf("public profile does not contain derived age: %s", data)
	}
}
