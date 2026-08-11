package user

import (
	"backend/api"
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
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	result, err := h.service.SaveProfileDecision(r.Context(), userID, targetUserID, req.Decision)
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
	case errors.Is(err, UserErrors.ProfilePictureRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Printf("Save profile decision from user %d to user %d: %v", userID, targetUserID, err)
		http.Error(w, "Failed to save profile decision", http.StatusInternalServerError)
		return
	}
	if result.DecisionChanged && result.ProfileDecision.Decision == models.ProfileDecisionLike && h.notifier != nil {
		_, notifyErr := h.notifier.NotifyLike(r.Context(), targetUserID, userID)
		if notifyErr != nil {
			log.Printf("Create like notification for user %d from user %d: %v", targetUserID, userID, notifyErr)
		}
	}
	if result.DecisionChanged && result.IsMatch && h.notifier != nil {
		_, notifyErr := h.notifier.NotifyMatch(r.Context(), targetUserID, userID)
		if notifyErr != nil {
			log.Printf("Create match notification for user %d from user %d: %v", targetUserID, userID, notifyErr)
		}
	}
	if result.MatchEnded && h.notifier != nil {
		_, notifyErr := h.notifier.NotifyUnlike(r.Context(), targetUserID, userID)
		if notifyErr != nil {
			log.Printf("Create unlike notification for user %d from user %d: %v", targetUserID, userID, notifyErr)
		}
	}
	response := SaveProfileDecisionResponse{
		ProfileDecision: result.ProfileDecision,
		IsMatch:         result.IsMatch,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile decision from user %d to user %d: %v", userID, targetUserID, err)
	}
}
