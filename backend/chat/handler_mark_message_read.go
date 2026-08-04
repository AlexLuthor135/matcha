package chat

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

type MarkMessageResponse struct {
	MessageID uint      `json:"message_id"`
	ReadAt    time.Time `json:"read_at"`
}

func (h *ChatHandler) MarkMessageRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawMessageID := r.PathValue("messageID")
	messageID, err := strconv.ParseUint(rawMessageID, 10, 64)
	if err != nil || messageID < 1 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}
	receipt, err := h.service.MarkMessageRead(r.Context(), userID, uint(messageID))
	if errors.Is(err, ChatErrors.MessageNotFound) {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Mark message %d for user %d: %v", messageID, userID, err)
		http.Error(w, "Failed to mark message as read", http.StatusInternalServerError)
		return
	}
	response := MarkMessageResponse{
		MessageID: receipt.MessageID,
		ReadAt:    receipt.ReadAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode read status for message %d: %v", messageID, err)
	}
}
