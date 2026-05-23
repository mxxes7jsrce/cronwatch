package jobregistry

import (
	"encoding/json"
	"net/http"
	"time"
)

// entryJSON is the JSON-serialisable view of an Entry.
type entryJSON struct {
	Name          string    `json:"name"`
	Schedule      string    `json:"schedule"`
	ExpectedEvery string    `json:"expected_every"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
	RegisteredAt  time.Time `json:"registered_at"`
	Stale         bool      `json:"stale"`
}

// Handler returns an http.HandlerFunc that serves the current registry
// snapshot as JSON, annotating each entry with its staleness.
func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		now := r.now()
		r.mu.RLock()
		out := make([]entryJSON, 0, len(r.entries))
		for _, e := range r.entries {
			stale := !e.LastSeen.IsZero() && now.Sub(e.LastSeen) > e.ExpectedEvery
			out = append(out, entryJSON{
				Name:          e.Name,
				Schedule:      e.Schedule,
				ExpectedEvery: e.ExpectedEvery.String(),
				LastSeen:      e.LastSeen,
				RegisteredAt:  e.RegisteredAt,
				Stale:         stale,
			})
		}
		r.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
