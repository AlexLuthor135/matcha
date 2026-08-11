package profile

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type ProfilePhotoResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

type ProfileResponse struct {
	ID                uint                   `json:"id"`
	UserName          string                 `json:"user_name"`
	FirstName         string                 `json:"first_name"`
	LastName          string                 `json:"last_name"`
	Email             string                 `json:"email"`
	Avatar            string                 `json:"avatar"`
	Gender            string                 `json:"gender"`
	Preferences       string                 `json:"preferences"`
	Bio               string                 `json:"bio"`
	Interests         []string               `json:"interests"`
	Photos            []ProfilePhotoResponse `json:"photos"`
	BirthDate         *string                `json:"birth_date"`
	Latitude          *float64               `json:"latitude"`
	Longitude         *float64               `json:"longitude"`
	LocationSource    models.LocationSource  `json:"location_source"`
	LocationName      string                 `json:"location_name"`
	LocationConsentAt *time.Time             `json:"location_consent_at"`
	FameRating        int64                  `json:"fame_rating"`
}

func profilePhotoResponses(photos []models.Photo) []ProfilePhotoResponse {
	response := make([]ProfilePhotoResponse, 0, len(photos))
	for _, photo := range photos {
		response = append(response, ProfilePhotoResponse{ID: photo.ID, URL: photo.URL})
	}
	return response
}

func profileBirthDateResponse(birthDate *time.Time) *string {
	if birthDate == nil {
		return nil
	}
	formatted := birthDate.Format(time.DateOnly)
	return &formatted
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	profile, err := h.service.GetProfile(r.Context(), userID)
	if errors.Is(err, ProfileErrors.UserNotFound) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Get profile for user %d: %v", userID, err)
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}
	response := ProfileResponse{
		ID:                profile.ID,
		UserName:          profile.UserName,
		FirstName:         profile.FirstName,
		LastName:          profile.LastName,
		Email:             profile.Email,
		Avatar:            profile.Avatar,
		Gender:            profile.Gender,
		Preferences:       profile.Preferences,
		Bio:               profile.Bio,
		Interests:         profile.Interests,
		Photos:            profilePhotoResponses(profile.Photos),
		BirthDate:         profileBirthDateResponse(profile.BirthDate),
		Latitude:          profile.Latitude,
		Longitude:         profile.Longitude,
		LocationSource:    profile.LocationSource,
		LocationName:      profile.LocationName,
		LocationConsentAt: profile.LocationConsentAt,
		FameRating:        profile.FameRating,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile for user %d: %v", userID, err)
	}
}
