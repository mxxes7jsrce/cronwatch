// Package healthcheck provides an HTTP endpoint for liveness probing of the
// cronwatch daemon. It exposes a simple /healthz route that returns the
// current uptime and a summary of monitored job states.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/cronwatch/internal/state"
)

// Response is the JSON body returned by the health endpoint.
type Response struct {
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	CheckedAt time.Time `json:"checked_at"`
	JobCount  int       `json:"job_count"`
	Missed    int       `json:"missed_jobs"`
	Failed    int       `json:"failed_jobs"`
}

// Server wraps an HTTP server that serves the health endpoint.
type Server struct {
	store   *state.Store
	started time.Time
	addr    string
}

// New creates a new health check Server listening on addr.
func New(addr string, store *state.Store) *Server {
	return &Server{
		store:   store,
		started: time.Now(),
		addr:    addr,
	}
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	snap := s.store.Snapshot()

	var missed, failed int
	for _, entry := range snap {
		if entry.MissedCount > 0 {
			missed++
		}
		if entry.FailureCount > 0 {
			failed++
		}
	}

	resp := Response{
		Status:    "ok",
		Uptime:    time.Since(s.started).Round(time.Second).String(),
		CheckedAt: time.Now().UTC(),
		JobCount:  len(snap),
		Missed:    missed,
		Failed:    failed,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
