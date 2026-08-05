package realtime

import (
	"backend/chat"
	"backend/middleware"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const maxIncomingEventSize = 16 << 10

type incomingEvent struct {
	Type        string `json:"type"`
	RecipientID uint   `json:"recipient_id"`
	MessageID   uint   `json:"message_id"`
	Message     string `json:"message"`
}

type chatMessageEvent struct {
	Type           string    `json:"type"`
	ID             uint      `json:"id"`
	ConversationID uint      `json:"conversation_id"`
	SenderID       uint      `json:"sender_id"`
	RecipientID    uint      `json:"recipient_id"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}

type messageReadEvent struct {
	Type           string    `json:"type"`
	MessageID      uint      `json:"message_id"`
	ConversationID uint      `json:"conversation_id"`
	ReaderID       uint      `json:"reader_id"`
	ReadAt         time.Time `json:"read_at"`
}

type errorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type WebsocketHandler struct {
	hub         *Hub
	chatService *chat.Service
	upgrader    websocket.Upgrader
}

func NewWebsocketHandler(hub *Hub, chatService *chat.Service) *WebsocketHandler {
	return &WebsocketHandler{
		hub:         hub,
		chatService: chatService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     isAllowedOrigin,
		},
	}
}

func isAllowedOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}
	requestHost := r.Header.Get("X-Forwarded-Host")
	if requestHost == "" {
		requestHost = r.Host
	}
	return strings.EqualFold(originURL.Host, requestHost)
}

func (h *WebsocketHandler) Connect(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	connection, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade WebSocket for user %d: %v", userID, err)
		return
	}
	client := newClient(userID, connection)
	h.hub.Register(client)
	defer h.hub.Unregister(client)
	go client.writeMessages()
	h.readMessages(r.Context(), client)
}

func (h *WebsocketHandler) readMessages(ctx context.Context, client *Client) {
	client.connection.SetReadLimit(maxIncomingEventSize)
	for {
		messageType, data, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Printf("Read error for user %d: %v", client.userID, err)
			}
			return
		}
		if messageType != websocket.TextMessage {
			h.sendError(client, "Only text messages are supported")
			continue
		}
		var event incomingEvent
		if err := json.Unmarshal(data, &event); err != nil {
			h.sendError(client, "Invalid WebSocket payload")
			continue
		}
		switch event.Type {
		case "chat_message":
			h.handleChatMessage(ctx, client, event)
		case "message_read":
			h.handleMessageRead(ctx, client, event)
		default:
			h.sendError(client, "Unknown event type")
		}
	}
}

func (h *WebsocketHandler) handleChatMessage(ctx context.Context, client *Client, event incomingEvent) {
	message, err := h.chatService.SaveMessage(ctx, client.userID, event.RecipientID, event.Message)
	if err != nil {
		switch {
		case errors.Is(err, chat.ChatErrors.CannotMessageSelf),
			errors.Is(err, chat.ChatErrors.MessageBlank),
			errors.Is(err, chat.ChatErrors.MessageTooLong),
			errors.Is(err, chat.ChatErrors.SenderNotFound),
			errors.Is(err, chat.ChatErrors.RecipientNotFound),
			errors.Is(err, chat.ChatErrors.UsersNotMatched):
			h.sendError(client, err.Error())
		default:
			log.Printf("Save message from user %d to user %d: %v", client.userID, event.RecipientID, err)
			h.sendError(client, "Failed to send message")
		}
		return
	}
	outgoing := chatMessageEvent{
		Type:           "chat_message",
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		RecipientID:    message.RecipientID,
		Message:        message.Content,
		CreatedAt:      message.CreatedAt,
	}
	data, err := json.Marshal(outgoing)
	if err != nil {
		log.Printf("Encode message %d: %v", message.ID, err)
		h.sendError(client, "Failed to send message")
		return
	}
	h.hub.SendToUser(message.SenderID, data)
	h.hub.SendToUser(message.RecipientID, data)
}

func (h *WebsocketHandler) handleMessageRead(ctx context.Context, client *Client, event incomingEvent) {
	if event.MessageID == 0 {
		h.sendError(client, "Invalid message ID")
		return
	}
	receipt, err := h.chatService.MarkMessageRead(ctx, client.userID, event.MessageID)
	if errors.Is(err, chat.ChatErrors.MessageNotFound) {
		h.sendError(client, "Message not found")
		return
	}
	if err != nil {
		log.Printf("Mark message %d as read for user %d: %v", event.MessageID, client.userID, err)
		h.sendError(client, "Failed to mark message as read")
		return
	}
	outgoing := messageReadEvent{
		Type:           "message_read",
		MessageID:      receipt.MessageID,
		ConversationID: receipt.ConversationID,
		ReaderID:       receipt.RecipientID,
		ReadAt:         receipt.ReadAt,
	}
	data, err := json.Marshal(outgoing)
	if err != nil {
		log.Printf("Encode read event for message %d: %v", receipt.MessageID, err)
		h.sendError(client, "Failed to mark message as read")
		return
	}
	h.hub.SendToUser(receipt.SenderID, data)
	h.hub.SendToUser(receipt.RecipientID, data)
}

func (h *WebsocketHandler) sendError(client *Client, message string) {
	data, err := json.Marshal(errorEvent{
		Type:    "error",
		Message: message,
	})
	if err != nil {
		log.Printf("Encode socket error for client %d: %v", client.userID, err)
		return
	}
	h.hub.SendToUser(client.userID, data)
}
