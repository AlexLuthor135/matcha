package realtime

import (
	"context"
	"log"
	"time"
)

type delivery struct {
	userID  uint
	message []byte
}

const lastSeenWriteTimeout = 5 * time.Second

type PresenceRecorder interface {
	RecordLastSeen(ctx context.Context, userID uint, lastSeenAt time.Time) error
}

type onlineRequest struct {
	userID   uint
	response chan bool
}

type Hub struct {
	clients          map[uint]map[*Client]struct{}
	register         chan *Client
	unregister       chan *Client
	deliver          chan delivery
	onlineRequests   chan onlineRequest
	presenceRecorder PresenceRecorder
}

func NewHub() *Hub {
	return &Hub{
		clients:        make(map[uint]map[*Client]struct{}),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		deliver:        make(chan delivery),
		onlineRequests: make(chan onlineRequest),
	}
}

func (h *Hub) SetPresenceRecorder(recorder PresenceRecorder) {
	h.presenceRecorder = recorder
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) SendToUser(userID uint, message []byte) {
	h.deliver <- delivery{
		userID:  userID,
		message: message,
	}
}

func (h *Hub) IsUserOnline(ctx context.Context, userID uint) bool {
	response := make(chan bool, 1)
	request := onlineRequest{
		userID:   userID,
		response: response,
	}
	select {
	case h.onlineRequests <- request:
	case <-ctx.Done():
		return false
	}
	select {
	case isOnline := <-response:
		return isOnline
	case <-ctx.Done():
		return false
	}
}

func (h *Hub) removeClient(client *Client) {
	userClients, exists := h.clients[client.userID]
	if !exists {
		return
	}
	if _, exists := userClients[client]; !exists {
		return
	}
	delete(userClients, client)
	close(client.send)
	if client.connection != nil {
		_ = client.connection.Close()
	}
	if len(userClients) != 0 {
		return
	}
	delete(h.clients, client.userID)
	h.recordLastSeen(client.userID)
}

func (h *Hub) recordLastSeen(userID uint) {
	if h.presenceRecorder == nil {
		return
	}
	lastSeenAt := time.Now().UTC()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), lastSeenWriteTimeout)
		defer cancel()
		if err := h.presenceRecorder.RecordLastSeen(ctx, userID, lastSeenAt); err != nil {
			log.Printf("Record last seen for user %d: %v", userID, err)
		}
	}()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]struct{})
			}
			h.clients[client.userID][client] = struct{}{}
		case client := <-h.unregister:
			h.removeClient(client)
		case outgoing := <-h.deliver:
			userClients := h.clients[outgoing.userID]
			for client := range userClients {
				select {
				case client.send <- outgoing.message:
				default:
					h.removeClient(client)
				}
			}
		case request := <-h.onlineRequests:
			request.response <- len(h.clients[request.userID]) > 0
		}
	}
}
