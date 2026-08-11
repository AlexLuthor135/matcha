package relationship

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
	FameRating  int64                  `json:"fame_rating"`
	Age         *int                   `json:"age"`
	Distance    *float64               `json:"distance"`
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

func profileAge(birthDate *time.Time) *int {
	if birthDate == nil {
		return nil
	}
	age := ageAt(*birthDate)
	return &age
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
		FameRating:  profile.FameRating,
		Age:         profileAge(profile.BirthDate),
		Distance:    profile.Distance,
	}
}

func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
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
	case errors.Is(err, RelationshipErrors.InvalidTargetUserID) || errors.Is(err, RelationshipErrors.InvalidLocation):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, RelationshipErrors.TargetUserNotFound):
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
