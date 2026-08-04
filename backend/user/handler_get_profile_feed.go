package user

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type PublicProfileResponse struct {
	ID          uint                   `json:"id"`
	UserName    string                 `json:"user_name"`
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	Gender      string                 `json:"gender"`
	Preferences string                 `json:"preferences"`
	Bio         string                 `json:"bio"`
	Interests   []string               `json:"interests"`
	Avatar      string                 `json:"avatar"`
	Photos      []ProfilePhotoResponse `json:"photos"`
}

type ProfileFeedResponse struct {
	Profiles []PublicProfileResponse `json:"profiles"`
	Count    int                     `json:"count"`
}

func parseProfileFeedLimit(rawLimit string) (int, error) {
	if rawLimit == "" {
		return defaultProfileFeedLimit, nil
	}
	limit, err := strconv.Atoi(rawLimit)
	if err != nil {
		return 0, UserErrors.InvalidProfileFeedLimit
	}
	return limit, nil
}

func publicProfileResponse(profile models.User) PublicProfileResponse {
	return PublicProfileResponse{
		ID:          profile.ID,
		UserName:    profile.UserName,
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		Gender:      profile.Gender,
		Preferences: profile.Preferences,
		Bio:         profile.Bio,
		Interests:   profile.Interests,
		Avatar:      profile.Avatar,
		Photos:      profilePhotoResponses(profile.Photos),
	}
}

func publicProfileResponses(profiles []models.User) []PublicProfileResponse {
	response := make([]PublicProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		response = append(response, publicProfileResponse(profile))
	}
	return response
}

func (h *UserHandler) GetProfileFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limit, err := parseProfileFeedLimit(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	profiles, err := h.service.GetProfileFeed(r.Context(), userID, limit)
	switch {
	case errors.Is(err, UserErrors.InvalidProfileFeedLimit):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Get profile feed for user %d: %v", userID, err)
		http.Error(w, "Failed to get profile feed", http.StatusInternalServerError)
		return
	}
	response := ProfileFeedResponse{
		Profiles: publicProfileResponses(profiles),
		Count:    len(profiles),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile feed for %d: %v", userID, err)
	}
}
