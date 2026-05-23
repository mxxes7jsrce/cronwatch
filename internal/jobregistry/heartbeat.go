package jobregistry

import (
	"net/http"
	"time"
)

// HeartbeatHandler returns an HTTP handler that records a heartbeat (touch)
// for the named job. External cron wrappers can POST to this endpoint to
// signal successful execution.
//
// Route convention: POST /heartbeat/{job}
func (r *Registry) HeartbeatHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		jobName := req.PathValue("job")
		if jobName == "" {
			http.Error(w, "missing job name", http.StatusBadRequest)
			return
		}

		if _, ok := r.Get(jobName); !ok {
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}

		r.Touch(jobName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","job":"` + jobName + `","ts":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
	})
}
