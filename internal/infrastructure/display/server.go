package display

import (
	"context"
	"net/http"

	"judo-app/internal/application/ports"
)

// Server serves the public display scoreboard over HTTP + WebSocket.
type Server struct {
	hub     *Hub
	handler *Handler
	addr    string
}

// NewServer creates a new display Server.
func NewServer(addr string) *Server {
	hub := NewHub()
	return &Server{
		hub:     hub,
		handler: NewHandler(hub),
		addr:    addr,
	}
}

// Broadcast implements ports.EventBroadcaster so the server can be passed to services.
func (s *Server) Broadcast(event ports.DisplayEvent) {
	s.hub.Broadcast(event)
}

// Start runs the HTTP server in a goroutine. Cancel ctx to stop.
func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handler.ServeWS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Display clients connect via WebSocket; this endpoint is a health check.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("judo-app display server"))
	})

	srv := &http.Server{Addr: s.addr, Handler: mux}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	go func() {
		_ = srv.ListenAndServe()
	}()
}
