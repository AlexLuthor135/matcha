package user

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type CompleteProfileRequest struct {
	Gender      string   `json:"gender"`
	Preferences string   `json:"preferences"`
	Bio         string   `json:"bio"`
	Interests   []string `json:"interests"`
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
		},
	)
	switch {
	case errors.Is(err, UserErrors.ProfileFieldsMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidGenderPreference):
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
