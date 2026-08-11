package profile

import (
	"backend/api"
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

type UpdateProfileRequest struct {
	Gender      *string          `json:"gender"`
	Preferences *string          `json:"preferences"`
	Bio         *string          `json:"bio"`
	Interests   *[]string        `json:"interests"`
	BirthDate   *string          `json:"birth_date"`
	Location    *LocationRequest `json:"location"`
}

type UpdateProfileResponse struct {
	Message string `json:"message"`
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdateProfileRequest
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	var birthDate *time.Time
	if req.BirthDate != nil {
		parsedBirthDate, err := time.Parse(time.DateOnly, strings.TrimSpace(*req.BirthDate))
		if err != nil {
			http.Error(w, "Birth date must use YYYY-MM-DD format", http.StatusBadRequest)
			return
		}
		birthDate = &parsedBirthDate
	}
	err := h.service.UpdateProfile(
		r.Context(),
		userID,
		UpdateProfileInput{
			Gender:      req.Gender,
			Preferences: req.Preferences,
			Bio:         req.Bio,
			Interests:   req.Interests,
			BirthDate:   birthDate,
			Location:    locationInput(req.Location),
		})
	switch {
	case errors.Is(err, ProfileErrors.NoProfileFields),
		errors.Is(err, ProfileErrors.InvalidGenderPreference),
		errors.Is(err, ProfileErrors.ProfileBioBlank),
		errors.Is(err, ProfileErrors.ProfileInterestsMissing),
		errors.Is(err, ProfileErrors.UserUnderage),
		errors.Is(err, ProfileErrors.InvalidLocation),
		errors.Is(err, ProfileErrors.InvalidLocationSource),
		errors.Is(err, ProfileErrors.LocationConsentRequired),
		errors.Is(err, ProfileErrors.InvalidInterestTag),
		errors.Is(err, ProfileErrors.ManualLocationNameMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, ProfileErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Update profile for user %d: %v", userID, err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	response := UpdateProfileResponse{
		Message: "Profile updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode updated profile for %d: %v", userID, err)
	}
}
