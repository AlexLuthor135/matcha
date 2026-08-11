package user

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"log"
	"net/http"
)

type ListProfileLikersResponse struct {
	Likers []models.ProfileLiker `json:"likers"`
}

func (h *UserHandler) ListProfileLikers(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	likers, err := h.service.ListProfileLikers(r.Context(), userID)
	if err != nil {
		log.Printf("List profile likers for user %d: %v", userID, err)
		http.Error(w, "Failed to get profile likers", http.StatusInternalServerError)
		return
	}
	response := ListProfileLikersResponse{Likers: likers}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile likers for user %d: %v", userID, err)
	}
}
