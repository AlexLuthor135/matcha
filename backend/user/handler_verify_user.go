package user

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type VerifyUserResponse struct {
	ID          uint `json:"id"`
	IsCompleted bool `json:"is_completed"`
}

func (h *UserHandler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isCompleted, err := h.service.VerifyUser(r.Context(), userID)
	if errors.Is(err, UserErrors.UserNotFound) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Verify user %d database error %v", userID, err)
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}
	response := VerifyUserResponse{
		ID:          userID,
		IsCompleted: isCompleted,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode verification for user %d: %v", userID, err)
	}
}
