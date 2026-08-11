package user

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"log"
	"net/http"
)

type ListProfileViewers struct {
	Viewers []models.ProfileViewer `json:"viewers"`
}

func (h *UserHandler) ListProfileViewers(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	viewers, err := h.service.ListProfileViewers(r.Context(), userID)
	if err != nil {
		log.Printf("List profile viewers for user %d: %v", userID, err)
		http.Error(w, "Failed to get profile viewers", http.StatusInternalServerError)
		return
	}
	response := ListProfileViewers{Viewers: viewers}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile viewers for user %d: %v", userID, err)
	}
}
