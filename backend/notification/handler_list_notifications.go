package notification

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"log"
	"net/http"
)

type ListNotificationsResponse struct {
	Notifications []models.Notification `json:"notifications"`
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	notifications, err := h.service.ListNotifications(r.Context(), userID)
	if err != nil {
		log.Printf("List notifications for user %d: %v", userID, err)
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}
	response := ListNotificationsResponse{
		Notifications: notifications,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode notifications for user %d: %v", userID, err)
	}
}
