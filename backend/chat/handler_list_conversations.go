package chat

import (
	"backend/middleware"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ConversationResponse struct {
	ID        uint      `json:"id"`
	UserOneID uint      `json:"user_one_id"`
	UserTwoID uint      `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListConversationsResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
}

func (h *ChatHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	conversations, err := h.service.ListConversations(r.Context(), userID)
	if err != nil {
		log.Printf("List conversations for user %d: %v", userID, err)
		http.Error(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}
	conversationsResponse := make([]ConversationResponse, 0, len(conversations))
	for _, conversation := range conversations {
		conversationsResponse = append(conversationsResponse, ConversationResponse{
			ID:        conversation.ID,
			UserOneID: conversation.UserOneID,
			UserTwoID: conversation.UserTwoID,
			CreatedAt: conversation.CreatedAt,
			UpdatedAt: conversation.UpdatedAt,
		})
	}
	response := ListConversationsResponse{
		Conversations: conversationsResponse,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode conversations for user %d: %v", userID, err)
	}
}
