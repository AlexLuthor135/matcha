package user

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type UpdateProfileRequest struct {
	Gender      *string   `json:"gender"`
	Preferences *string   `json:"preferences"`
	Bio         *string   `json:"bio"`
	Interests   *[]string `json:"interests"`
}

type UpdateProfileResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err := h.service.UpdateProfile(
		r.Context(),
		userID,
		UpdateProfileInput{
			Gender:      req.Gender,
			Preferences: req.Preferences,
			Bio:         req.Bio,
			Interests:   req.Interests,
		})
	switch {
	case errors.Is(err, UserErrors.NoProfileFields):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidGenderPreference):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.ProfileBioBlank):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.ProfileInterestsMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
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
