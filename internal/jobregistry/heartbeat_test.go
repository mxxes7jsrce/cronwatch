package jobregistry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

func heartbeatJobs() []config.Job {
	return []config.Job{
		{Name: "backup", Schedule: "@daily", LogPath: "/var/log/backup.log", Timeout: 2 * time.Hour},
	}
}

func TestHeartbeat_RecordsTouchOnPost(t *testing.T) {
	r := newWithClock(heartbeatJobs(), fixedNow)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/backup", nil)
	req.SetPathValue("job", "backup")
	w := httptest.NewRecorder()

	r.HeartbeatHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"job":"backup"`) {
		t.Errorf("expected job name in response, got: %s", body)
	}
}

func TestHeartbeat_UnknownJobReturns404(t *testing.T) {
	r := newWithClock(heartbeatJobs(), fixedNow)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/ghost", nil)
	req.SetPathValue("job", "ghost")
	w := httptest.NewRecorder()

	r.HeartbeatHandler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHeartbeat_MethodNotAllowedOnGet(t *testing.T) {
	r := newWithClock(heartbeatJobs(), fixedNow)

	req := httptest.NewRequest(http.MethodGet, "/heartbeat/backup", nil)
	req.SetPathValue("job", "backup")
	w := httptest.NewRecorder()

	r.HeartbeatHandler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHeartbeat_MissingJobNameReturns400(t *testing.T) {
	r := newWithClock(heartbeatJobs(), fixedNow)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/", nil)
	// PathValue intentionally not set — simulates missing segment
	w := httptest.NewRecorder()

	r.HeartbeatHandler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHeartbeat_ContentTypeJSON(t *testing.T) {
	r := newWithClock(heartbeatJobs(), fixedNow)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/backup", nil)
	req.SetPathValue("job", "backup")
	w := httptest.NewRecorder()

	r.HeartbeatHandler().ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
