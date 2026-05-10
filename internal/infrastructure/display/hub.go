package display

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected display browser.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub manages all connected display WebSocket clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	close(c.send)
}

// Broadcast sends a payload to all connected clients.
func (h *Hub) Broadcast(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			// Slow client — drop the message to avoid blocking.
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
