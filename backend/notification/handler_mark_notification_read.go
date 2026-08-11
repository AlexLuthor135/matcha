package notification

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

type MarkNotificationResponse struct {
	NotificationID uint      `json:"notification_id"`
	ReadAt         time.Time `json:"read_at"`
}

func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawNotificationID := r.PathValue("notificationID")
	notificationID, err := strconv.ParseUint(rawNotificationID, 10, strconv.IntSize)
	if err != nil || notificationID < 1 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	receipt, err := h.service.MarkNotificationRead(r.Context(), userID, uint(notificationID))
	if errors.Is(err, NotificationErrors.NotificationNotFound) {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Mark notification %d for user %d: %v", notificationID, userID, err)
		http.Error(w, "Failed to mark notification read", http.StatusInternalServerError)
		return
	}
	response := MarkNotificationResponse{
		NotificationID: receipt.NotificationID,
		ReadAt:         receipt.ReadAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode notification %d read status: %v", notificationID, err)
	}

}
