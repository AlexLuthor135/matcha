package user

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ProfilePhotoResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

type ProfileResponse struct {
	ID          uint                   `json:"id"`
	UserName    string                 `json:"user_name"`
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	Email       string                 `json:"email"`
	Avatar      string                 `json:"avatar"`
	Gender      string                 `json:"gender"`
	Preferences string                 `json:"preferences"`
	Bio         string                 `json:"bio"`
	Interests   []string               `json:"interests"`
	Photos      []ProfilePhotoResponse `json:"photos"`
}

func profilePhotoResponses(photos []models.Photo) []ProfilePhotoResponse {
	response := make([]ProfilePhotoResponse, 0, len(photos))
	for _, photo := range photos {
		response = append(response, ProfilePhotoResponse{ID: photo.ID, URL: photo.URL})
	}
	return response
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	profile, err := h.service.GetProfile(r.Context(), userID)
	if errors.Is(err, UserErrors.UserNotFound) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Get profile for user %d: %v", userID, err)
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}
	response := ProfileResponse{
		ID:          profile.ID,
		UserName:    profile.UserName,
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		Email:       profile.Email,
		Avatar:      profile.Avatar,
		Gender:      profile.Gender,
		Preferences: profile.Preferences,
		Bio:         profile.Bio,
		Interests:   profile.Interests,
		Photos:      profilePhotoResponses(profile.Photos),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile for user %d: %v", userID, err)
	}
}
