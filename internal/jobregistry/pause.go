package jobregistry

import (
	"encoding/json"
	"net/http"
)

// PauseHandler returns an http.Handler that allows pausing and resuming
// jobs in the registry. A paused job will not trigger missed-run alerts.
//
// POST /pause?job=<name>   — pause a job
// POST /resume?job=<name>  — resume a paused job
// GET  /pause              — list all paused jobs
func (r *Registry) PauseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			r.listPaused(w)
		case http.MethodPost:
			action := req.URL.Query().Get("action")
			name := req.URL.Query().Get("job")
			if name == "" {
				http.Error(w, "missing job query parameter", http.StatusBadRequest)
				return
			}
			switch action {
			case "pause":
				r.setPaused(w, name, true)
			case "resume":
				r.setPaused(w, name, false)
			default:
				http.Error(w, "action must be 'pause' or 'resume'", http.StatusBadRequest)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func (r *Registry) listPaused(w http.ResponseWriter) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paused := []string{}
	for name, entry := range r.entries {
		if entry.Paused {
			paused = append(paused, name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"paused": paused})
}

func (r *Registry) setPaused(w http.ResponseWriter, name string, paused bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[name]
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	entry.Paused = paused
	r.entries[name] = entry

	action := "paused"
	if !paused {
		action = "resumed"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"job": name, "status": action})
}
