package realtime

import "github.com/gorilla/websocket"

const clientSendBufferSize = 64

type Client struct {
	userID     uint
	connection *websocket.Conn
	send       chan []byte
}

func newClient(userID uint, connection *websocket.Conn) *Client {
	return &Client{
		userID:     userID,
		connection: connection,
		send:       make(chan []byte, clientSendBufferSize),
	}
}

func (c *Client) writeMessages() {
	defer c.connection.Close()
	for message := range c.send {
		if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
