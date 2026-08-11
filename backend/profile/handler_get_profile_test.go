package profile

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
	handler := NewHandler(NewService(repository, &fakeImageStorage{}))
	request := authenticatedUserRequest(httptest.NewRequest(http.MethodGet, "/profile", nil), userID)
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
	if body.Latitude == nil || *body.Latitude != latitude || body.Longitude == nil || *body.Longitude != longitude {
		t.Fatalf("coordinates = (%v, %v), want (%v, %v)", body.Latitude, body.Longitude, latitude, longitude)
	}
	if body.LocationSource != models.LocationSourceGPS || body.LocationConsentAt == nil || !body.LocationConsentAt.Equal(consentAt) {
		t.Fatalf("location metadata = %+v", body)
	}
}
