package user

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
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

type PublicProfileDetailsResponse struct {
	PublicProfileResponse
	IsOnline    bool      `json:"is_online"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	LikedByMe   bool      `json:"liked_by_me"`
	LikedMe     bool      `json:"liked_me"`
	IsConnected bool      `json:"is_connected"`
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

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawTargetUserID := r.PathValue("targetUserID")
	parsedTargetUserID, err := strconv.ParseUint(rawTargetUserID, 10, strconv.IntSize)
	if err != nil || parsedTargetUserID == 0 {
		http.Error(w, "Invalid target user id", http.StatusBadRequest)
		return
	}
	targetUserID := uint(parsedTargetUserID)
	result, err := h.service.GetPublicProfile(r.Context(), viewerID, targetUserID)
	switch {
	case errors.Is(err, UserErrors.InvalidTargetUserID) || errors.Is(err, UserErrors.InvalidLocation):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.TargetUserNotFound):
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Get public profile %d for user %d: %v", targetUserID, viewerID, err)
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}
	profile := result.Profile
	if viewerID != targetUserID {
		_, err := h.service.RecordProfileView(r.Context(), viewerID, targetUserID)
		if err != nil {
			log.Printf("Record profile view from user %d to user %d: %v", viewerID, targetUserID, err)
			http.Error(w, "Failed to record profile view", http.StatusInternalServerError)
			return
		}
		if h.notifier != nil {
			_, notifyErr := h.notifier.NotifyProfileView(r.Context(), targetUserID, viewerID)
			if notifyErr != nil {
				log.Printf("Notify user %d about profile view from %d: %v", targetUserID, viewerID, notifyErr)
			}
		}
	}
	isOnline := false
	if h.presence != nil {
		isOnline = h.presence.IsUserOnline(r.Context(), targetUserID)
	}
	response := PublicProfileDetailsResponse{
		PublicProfileResponse: publicProfileResponse(profile),
		IsOnline:              isOnline,
		LastSeenAt:            profile.LastSeenAt,
		LikedByMe:             result.Relationship.LikedByMe,
		LikedMe:               result.Relationship.LikedMe,
		IsConnected:           result.Relationship.IsConnected(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode public profile %d for user %d: %v", targetUserID, viewerID, err)
	}
}
