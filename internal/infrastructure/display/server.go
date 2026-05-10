package display

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"judo-app/internal/application"
	"judo-app/internal/application/ports"
	"judo-app/internal/domain"

	"github.com/google/uuid"
)

// Server serves the public display scoreboard over HTTP + WebSocket,
// and the Angular SPA as static files so remote operators can connect
// via browser at http://<host>:8080.
type Server struct {
	hub     *Hub
	handler *Handler
	addr    string
	tatami  *application.TatamiService // set after construction via SetTatamiService
	spa     fs.FS                       // embedded Angular build; nil = no SPA
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

// SetTatamiService wires the tatami service so REST handlers can call it.
// Must be called before Start.
func (s *Server) SetTatamiService(t *application.TatamiService) {
	s.tatami = t
}

// SetSPA sets the embedded filesystem for serving the Angular SPA.
// Pass the sub-FS rooted at build/frontend/browser.
func (s *Server) SetSPA(f fs.FS) {
	s.spa = f
}

// Broadcast implements ports.EventBroadcaster so the server can be passed to services.
func (s *Server) Broadcast(event ports.DisplayEvent) {
	s.hub.Broadcast(event)
}

// Start runs the HTTP server in a goroutine. Cancel ctx to stop.
func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()

	// WebSocket endpoint.
	mux.HandleFunc("/ws", s.handler.ServeWS)

	// REST API for remote browser operators.
	mux.HandleFunc("/api/matches", s.handleListMatches)
	mux.HandleFunc("/api/matches/", s.handleMatchAction) // /api/matches/{id}/claim|result

	// SPA fallback — serves Angular static files; unknown paths → index.html.
	if s.spa != nil {
		mux.Handle("/", s.spaHandler())
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("judo-app display server"))
		})
	}

	srv := &http.Server{Addr: s.addr, Handler: mux}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	go func() {
		_ = srv.ListenAndServe()
	}()
}

// ── SPA handler ───────────────────────────────────────────────────────────────

// spaHandler returns an http.Handler that serves the embedded SPA.
// Requests to unknown paths fall back to index.html for client-side routing.
func (s *Server) spaHandler() http.Handler {
	fileServer := http.FileServer(http.FS(s.spa))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to open the file directly.
		f, err := s.spa.Open(strings.TrimPrefix(r.URL.Path, "/"))
		if err != nil {
			// Not found → serve index.html for SPA routing.
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		_ = f.Close()
		fileServer.ServeHTTP(w, r)
	})
}

// ── REST handlers ─────────────────────────────────────────────────────────────

// GET /api/matches?tournamentId=<uuid>
func (s *Server) handleListMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.tatami == nil {
		http.Error(w, "service not ready", http.StatusServiceUnavailable)
		return
	}
	tIDStr := r.URL.Query().Get("tournamentId")
	tID, err := uuid.Parse(tIDStr)
	if err != nil {
		http.Error(w, "invalid tournamentId", http.StatusBadRequest)
		return
	}
	rows, err := s.tatami.ListMatches(r.Context(), tID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

// POST /api/matches/{id}/claim   body: {"tatamiId":"1","labelA":"Uke","labelB":"Tori"}
// POST /api/matches/{id}/result  body: {"categoryId":"...","winnerIdx":0,"method":"IPPON"}
func (s *Server) handleMatchAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.tatami == nil {
		http.Error(w, "service not ready", http.StatusServiceUnavailable)
		return
	}

	// Path: /api/matches/{id}/claim  or  /api/matches/{id}/result
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/matches/"), "/")
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	matchIDStr, action := parts[0], parts[1]
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		http.Error(w, "invalid matchId", http.StatusBadRequest)
		return
	}

	switch action {
	case "claim":
		var body struct {
			TatamiID string `json:"tatamiId"`
			LabelA   string `json:"labelA"`
			LabelB   string `json:"labelB"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := s.tatami.ClaimMatch(r.Context(), matchID, body.TatamiID, body.LabelA, body.LabelB); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case "result":
		var body struct {
			CategoryID string `json:"categoryId"`
			WinnerIdx  int    `json:"winnerIdx"`
			Method     string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		catID, err := uuid.Parse(body.CategoryID)
		if err != nil {
			http.Error(w, "invalid categoryId", http.StatusBadRequest)
			return
		}
		if err := s.tatami.RecordResult(r.Context(), catID, matchID, body.WinnerIdx, domain.FinishMethod(body.Method)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}
