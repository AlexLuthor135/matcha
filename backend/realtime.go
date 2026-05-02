package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goji.io/pat"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultRealtimeLimit = 50
const maxRealtimeLimit = 100

type chatMessageResponse struct {
	ID             uint       `json:"id"`
	ConversationID uint       `json:"conversation_id"`
	SenderID       uint       `json:"sender_id"`
	RecipientID    uint       `json:"recipient_id"`
	Content        string     `json:"content"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type conversationResponse struct {
	ID          uint                 `json:"id"`
	UserOneID   uint                 `json:"user_one_id"`
	UserTwoID   uint                 `json:"user_two_id"`
	LastMessage *chatMessageResponse `json:"last_message,omitempty"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type notificationResponse struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	SenderID  *uint      `json:"sender_id,omitempty"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Data      any        `json:"data,omitempty"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func saveChatMessage(senderID uint, recipientID uint, content string) (ChatMessage, error) {
	content = strings.TrimSpace(content)
	if recipientID == 0 || content == "" {
		return ChatMessage{}, errors.New("recipient_id and message are required")
	}
	if senderID == recipientID {
		return ChatMessage{}, errors.New("cannot send a message to yourself")
	}

	var recipient User
	if err := DB.Select("id").First(&recipient, recipientID).Error; err != nil {
		return ChatMessage{}, err
	}

	var message ChatMessage
	err := DB.Transaction(func(tx *gorm.DB) error {
		conversation, err := findOrCreateConversation(tx, senderID, recipientID)
		if err != nil {
			return err
		}

		message = ChatMessage{
			ConversationID: conversation.ID,
			SenderID:       senderID,
			RecipientID:    recipientID,
			Content:        content,
		}

		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		return tx.Model(&conversation).Update("updated_at", message.CreatedAt).Error
	})

	return message, err
}

func findOrCreateConversation(db *gorm.DB, userAID uint, userBID uint) (Conversation, error) {
	userOneID, userTwoID := orderUserIDs(userAID, userBID)
	conversation := Conversation{
		UserOneID: userOneID,
		UserTwoID: userTwoID,
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_one_id"}, {Name: "user_two_id"}},
		DoNothing: true,
	}).Create(&conversation).Error
	if err != nil {
		return Conversation{}, err
	}

	err = db.Where("user_one_id = ? AND user_two_id = ?", userOneID, userTwoID).First(&conversation).Error
	return conversation, err
}

func orderUserIDs(userAID uint, userBID uint) (uint, uint) {
	if userAID < userBID {
		return userAID, userBID
	}
	return userBID, userAID
}

func saveNotification(userID uint, senderID *uint, title string, message string, data json.RawMessage) (Notification, error) {
	message = strings.TrimSpace(message)
	if userID == 0 || message == "" {
		return Notification{}, errors.New("user_id and message are required")
	}

	if len(data) == 0 {
		data = json.RawMessage("{}")
	}
	if !json.Valid(data) {
		return Notification{}, errors.New("data must be valid JSON")
	}

	var recipient User
	if err := DB.Select("id").First(&recipient, userID).Error; err != nil {
		return Notification{}, err
	}

	notification := Notification{
		UserID:   userID,
		SenderID: senderID,
		Type:     "notification",
		Title:    title,
		Message:  message,
		Data:     string(data),
	}

	if err := DB.Create(&notification).Error; err != nil {
		return Notification{}, err
	}

	return notification, nil
}

func createSystemNotification(hub *websocketHub, userID uint, title string, message string, data any) (Notification, error) {
	raw := json.RawMessage("{}")
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return Notification{}, err
		}
		raw = encoded
	}

	notification, err := saveNotification(userID, nil, title, message, raw)
	if err != nil {
		return Notification{}, err
	}

	hub.sendToUser(websocketEvent{
		Type:        "notification",
		ID:          notification.ID,
		RecipientID: userID,
		Title:       title,
		Message:     notification.Message,
		Data:        raw,
		CreatedAt:   notification.CreatedAt,
	})

	return notification, nil
}

func requireStaffUser(userID uint) error {
	var user User
	if err := DB.Select("id", "is_staff").First(&user, userID).Error; err != nil {
		return err
	}
	if !user.IsStaff {
		return errors.New("staff access required")
	}
	return nil
}

func listConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := parseRealtimeLimit(r.URL.Query().Get("limit"))
	var conversations []Conversation
	if err := DB.
		Where("user_one_id = ? OR user_two_id = ?", userID, userID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&conversations).Error; err != nil {
		http.Error(w, "Failed to load conversations", http.StatusInternalServerError)
		return
	}

	responses := make([]conversationResponse, 0, len(conversations))
	for _, conversation := range conversations {
		response := conversationResponse{
			ID:        conversation.ID,
			UserOneID: conversation.UserOneID,
			UserTwoID: conversation.UserTwoID,
			UpdatedAt: conversation.UpdatedAt,
		}

		var lastMessage ChatMessage
		if err := DB.Where("conversation_id = ?", conversation.ID).Order("created_at DESC").First(&lastMessage).Error; err == nil {
			messageResponse := buildChatMessageResponse(lastMessage)
			response.LastMessage = &messageResponse
		}

		responses = append(responses, response)
	}

	writeJSON(w, http.StatusOK, map[string]any{"conversations": responses})
}

func listConversationMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID, ok := parseUintRouteParam(r, "conversationID")
	if !ok {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	var conversation Conversation
	if err := DB.Where("id = ? AND (user_one_id = ? OR user_two_id = ?)", conversationID, userID, userID).First(&conversation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load conversation", http.StatusInternalServerError)
		return
	}

	limit := parseRealtimeLimit(r.URL.Query().Get("limit"))
	var messages []ChatMessage
	if err := DB.
		Where("conversation_id = ?", conversation.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}

	responses := make([]chatMessageResponse, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		responses = append(responses, buildChatMessageResponse(messages[i]))
	}

	writeJSON(w, http.StatusOK, map[string]any{"messages": responses})
}

func markMessageRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID, ok := parseUintRouteParam(r, "messageID")
	if !ok {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	var message ChatMessage
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND recipient_id = ?", messageID, userID).First(&message).Error; err != nil {
			return err
		}
		if message.ReadAt == nil {
			message.ReadAt = &now
			return tx.Model(&message).Update("read_at", now).Error
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Message not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to mark message as read", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, buildChatMessageResponse(message))
}

func listNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := parseRealtimeLimit(r.URL.Query().Get("limit"))
	query := DB.Where("user_id = ?", userID)
	if r.URL.Query().Get("unread") == "true" {
		query = query.Where("read_at IS NULL")
	}

	var notifications []Notification
	if err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error; err != nil {
		http.Error(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}

	responses := make([]notificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		responses = append(responses, buildNotificationResponse(notification))
	}

	writeJSON(w, http.StatusOK, map[string]any{"notifications": responses})
}

func markNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationID, ok := parseUintRouteParam(r, "notificationID")
	if !ok {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	var notification Notification
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
			return err
		}
		if notification.ReadAt == nil {
			notification.ReadAt = &now
			return tx.Model(&notification).Update("read_at", now).Error
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, buildNotificationResponse(notification))
}

func buildChatMessageResponse(message ChatMessage) chatMessageResponse {
	return chatMessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		RecipientID:    message.RecipientID,
		Content:        message.Content,
		ReadAt:         message.ReadAt,
		CreatedAt:      message.CreatedAt,
	}
}

func buildNotificationResponse(notification Notification) notificationResponse {
	var data any
	if notification.Data != "" {
		_ = json.Unmarshal([]byte(notification.Data), &data)
	}

	return notificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		SenderID:  notification.SenderID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Data:      data,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
	}
}

func parseRealtimeLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultRealtimeLimit
	}
	if limit > maxRealtimeLimit {
		return maxRealtimeLimit
	}
	return limit
}

func parseUintRouteParam(r *http.Request, name string) (uint, bool) {
	value := pat.Param(r, name)
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}
	return uint(parsed), true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
