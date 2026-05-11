package display

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins for local display
}

// Handler handles WebSocket upgrade requests from display clients.
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS handles the WebSocket upgrade and registers the client.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{
		conn: conn,
		send: make(chan []byte, 64),
	}
	h.hub.Register(client)
	go client.writePump()

	// Read pump — keep connection alive and detect close.
	defer h.hub.Unregister(client)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
