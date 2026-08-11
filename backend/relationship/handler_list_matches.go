package relationship

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"log"
	"net/http"
)

type ListMatchesResponse struct {
	Matches []models.Match `json:"matches"`
}

func (h *Handler) ListMatches(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	matches, err := h.service.ListMatches(r.Context(), userID)
	if err != nil {
		log.Printf("List matches for user %d: %v", userID, err)
		http.Error(w, "Failed to get matches", http.StatusInternalServerError)
		return
	}
	response := ListMatchesResponse{
		Matches: matches,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode matches for user %d: %v", userID, err)
	}
}
