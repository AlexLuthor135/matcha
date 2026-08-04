package realtime

type delivery struct {
	userID  uint
	message []byte
}

type Hub struct {
	clients    map[uint]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	deliver    chan delivery
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		deliver:    make(chan delivery),
	}
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

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]struct{})
			}
			h.clients[client.userID][client] = struct{}{}
		case client := <-h.unregister:
			userClients, exists := h.clients[client.userID]
			if !exists {
				continue
			}
			if _, exists := userClients[client]; !exists {
				continue
			}
			delete(userClients, client)
			close(client.send)
			if len(userClients) == 0 {
				delete(h.clients, client.userID)
			}
		case outgoing := <-h.deliver:
			userClients := h.clients[outgoing.userID]
			for client := range userClients {
				select {
				case client.send <- outgoing.message:
				default:
					delete(userClients, client)
					close(client.send)
					_ = client.connection.Close()
				}
			}
			if len(userClients) == 0 {
				delete(h.clients, outgoing.userID)
			}
		}
	}
}
