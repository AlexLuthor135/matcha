package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type websocketEvent struct {
	Type           string          `json:"type"`
	ID             uint            `json:"id,omitempty"`
	ConversationID uint            `json:"conversation_id,omitempty"`
	RecipientID    uint            `json:"recipient_id,omitempty"`
	SenderID       uint            `json:"sender_id,omitempty"`
	Message        string          `json:"message,omitempty"`
	Title          string          `json:"title,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	CreatedAt      time.Time       `json:"created_at,omitempty"`
}

type websocketClient struct {
	hub    *websocketHub
	userID uint
	conn   *websocket.Conn
	send   chan websocketEvent
}

type websocketHub struct {
	clients    map[uint]map[*websocketClient]bool
	register   chan *websocketClient
	unregister chan *websocketClient
	deliver    chan websocketEvent
}

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     isAllowedWebsocketOrigin,
}

func isAllowedWebsocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	allowedOrigin := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
	if allowedOrigin != "" {
		return origin == allowedOrigin
	}

	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := parsedOrigin.Hostname()
	return host == "localhost" || host == "127.0.0.1"
}

func newWebsocketHub() *websocketHub {
	return &websocketHub{
		clients:    make(map[uint]map[*websocketClient]bool),
		register:   make(chan *websocketClient),
		unregister: make(chan *websocketClient),
		deliver:    make(chan websocketEvent),
	}
}

func (hub *websocketHub) run() {
	for {
		select {
		case client := <-hub.register:
			if hub.clients[client.userID] == nil {
				hub.clients[client.userID] = make(map[*websocketClient]bool)
			}
			hub.clients[client.userID][client] = true

		case client := <-hub.unregister:
			if clients := hub.clients[client.userID]; clients != nil {
				if clients[client] {
					delete(clients, client)
					close(client.send)
				}
				if len(clients) == 0 {
					delete(hub.clients, client.userID)
				}
			}

		case event := <-hub.deliver:
			if event.CreatedAt.IsZero() {
				event.CreatedAt = time.Now().UTC()
			}

			clients := hub.clients[event.RecipientID]
			for client := range clients {
				select {
				case client.send <- event:
				default:
					delete(clients, client)
					close(client.send)
				}
			}

			if len(clients) == 0 {
				delete(hub.clients, event.RecipientID)
			}
		}
	}
}

func (hub *websocketHub) sendToUser(event websocketEvent) {
	hub.deliver <- event
}

func websocketHandler(hub *websocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade failed: %v", err)
			return
		}

		client := &websocketClient{
			hub:    hub,
			userID: userID,
			conn:   conn,
			send:   make(chan websocketEvent, 256),
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

func (client *websocketClient) readPump() {
	defer func() {
		client.hub.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var event websocketEvent
		if err := client.conn.ReadJSON(&event); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read failed: %v", err)
			}
			break
		}

		event.SenderID = client.userID
		event.CreatedAt = time.Now().UTC()

		switch event.Type {
		case "chat_message":
			if event.RecipientID == 0 || event.Message == "" {
				client.send <- websocketEvent{Type: "error", Message: "recipient_id and message are required"}
				continue
			}

			message, err := saveChatMessage(client.userID, event.RecipientID, event.Message)
			if err != nil {
				log.Printf("failed to save websocket message: %v", err)
				client.send <- websocketEvent{Type: "error", Message: "failed to save message"}
				continue
			}

			event.ID = message.ID
			event.ConversationID = message.ConversationID
			event.Message = message.Content
			event.CreatedAt = message.CreatedAt

			client.send <- event
			client.hub.sendToUser(event)

		default:
			client.send <- websocketEvent{Type: "error", Message: "unsupported websocket event type"}
		}
	}
}

func (client *websocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case event, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteJSON(event); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type notificationRequest struct {
	UserID  uint            `json:"user_id"`
	Title   string          `json:"title"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func notificationHandler(hub *websocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := r.Context().Value(userIDKey).(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := requireStaffUser(currentUserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var req notificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.UserID == 0 || req.Message == "" {
			http.Error(w, "user_id and message are required", http.StatusBadRequest)
			return
		}
		if len(req.Data) > 0 && !json.Valid(req.Data) {
			http.Error(w, "data must be valid JSON", http.StatusBadRequest)
			return
		}

		notification, err := saveNotification(req.UserID, &currentUserID, req.Title, req.Message, req.Data)
		if err != nil {
			log.Printf("failed to save notification: %v", err)
			http.Error(w, "Failed to save notification", http.StatusInternalServerError)
			return
		}

		hub.sendToUser(websocketEvent{
			Type:        "notification",
			ID:          notification.ID,
			RecipientID: req.UserID,
			SenderID:    currentUserID,
			Title:       req.Title,
			Message:     req.Message,
			Data:        req.Data,
			CreatedAt:   notification.CreatedAt,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"message": "notification queued"})
	}
}
