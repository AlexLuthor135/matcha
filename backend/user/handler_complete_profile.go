package user

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

type CompleteProfileRequest struct {
	Gender      string           `json:"gender"`
	Preferences string           `json:"preferences"`
	Bio         string           `json:"bio"`
	Interests   []string         `json:"interests"`
	BirthDate   string           `json:"birth_date"`
	Location    *LocationRequest `json:"location"`
}

type CompleteProfileResponse struct {
	Message     string `json:"message"`
	IsCompleted bool   `json:"is_completed"`
}

func (h *UserHandler) CompleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req CompleteProfileRequest
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	birthDate, err := time.Parse(time.DateOnly, strings.TrimSpace(req.BirthDate))
	if err != nil {
		http.Error(w, "Birth date must use YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	isCompleted, err := h.service.CompleteProfile(
		r.Context(),
		userID,
		CompleteProfileInput{
			Gender:      req.Gender,
			Preferences: req.Preferences,
			Bio:         req.Bio,
			Interests:   req.Interests,
			BirthDate:   birthDate,
			Location:    locationInput(req.Location),
		},
	)
	switch {
	case errors.Is(err, UserErrors.ProfileFieldsMissing) ||
		errors.Is(err, UserErrors.InvalidGenderPreference) ||
		errors.Is(err, UserErrors.UserUnderage) ||
		errors.Is(err, UserErrors.InvalidLocation) ||
		errors.Is(err, UserErrors.InvalidLocationSource) ||
		errors.Is(err, UserErrors.LocationConsentRequired) ||
		errors.Is(err, UserErrors.ManualLocationNameMissing) ||
		errors.Is(err, UserErrors.InvalidInterestTag) ||
		errors.Is(err, UserErrors.ProfilePictureRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Complete profile for user %d: %v", userID, err)
		http.Error(w, "Failed to complete profile", http.StatusInternalServerError)
		return
	}
	response := CompleteProfileResponse{
		Message:     "Profile completed successfully",
		IsCompleted: isCompleted,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode complete profile for %d: %v", userID, err)
	}
}
