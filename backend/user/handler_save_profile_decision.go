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

type SaveProfileDecisionRequest struct {
	Decision models.ProfileDecisionValue `json:"decision"`
}

type SaveProfileDecisionResponse struct {
	ProfileDecision models.ProfileDecision `json:"profile_decision"`
	IsMatch         bool                   `json:"is_match"`
}

func (h *UserHandler) SaveProfileDecision(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawTargetUserID := r.PathValue("targetUserID")
	parsedTargetUserID, err := strconv.ParseUint(rawTargetUserID, 10, strconv.IntSize)
	if err != nil || parsedTargetUserID == 0 {
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}
	targetUserID := uint(parsedTargetUserID)
	var req SaveProfileDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	savedDecision, isMatch, err := h.service.SaveProfileDecision(r.Context(), userID, targetUserID, req.Decision)
	switch {
	case errors.Is(err, UserErrors.InvalidTargetUserID):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.CannotDecideOwnProfile):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidProfileDecision):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case errors.Is(err, UserErrors.TargetUserNotFound):
		http.Error(w, "Target user not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Save profile decision from user %d to user %d: %v", userID, targetUserID, err)
		http.Error(w, "Failed to save profile decision", http.StatusInternalServerError)
		return
	}
	response := SaveProfileDecisionResponse{
		ProfileDecision: savedDecision,
		IsMatch:         isMatch,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile decision from user %d to user %d: %v", userID, targetUserID, err)
	}
}
