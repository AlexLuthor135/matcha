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

type MessageResponse struct {
	ID             uint       `json:"id"`
	ConversationID uint       `json:"conversation_id"`
	SenderID       uint       `json:"sender_id"`
	RecipientID    uint       `json:"recipient_id"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ReadAt         *time.Time `json:"read_at"`
}

type ListConversationMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

func (h *ChatHandler) ListConversationMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawConversationID := r.PathValue("conversationID")
	conversationID, err := strconv.ParseUint(rawConversationID, 10, 64)
	if err != nil || conversationID < 1 {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	messages, err := h.service.ListConversationMessages(r.Context(), userID, uint(conversationID))
	if errors.Is(err, ChatErrors.ConversationNotFound) {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("List messages for conversation %d: %v", conversationID, err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}
	messageResponses := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		messageResponses = append(messageResponses, MessageResponse{
			ID:             message.ID,
			ConversationID: message.ConversationID,
			SenderID:       message.SenderID,
			RecipientID:    message.RecipientID,
			Content:        message.Content,
			CreatedAt:      message.CreatedAt,
			UpdatedAt:      message.UpdatedAt,
			ReadAt:         message.ReadAt,
		})
	}
	response := ListConversationMessagesResponse{
		Messages: messageResponses,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode messages for conversation %d: %v", conversationID, err)
	}
}
