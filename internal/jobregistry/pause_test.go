package jobregistry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/config"
)

func pauseJobs() []config.Job {
	return []config.Job{
		{Name: "backup", Schedule: "@daily", LogPath: "/var/log/backup.log"},
		{Name: "report", Schedule: "@weekly", LogPath: "/var/log/report.log"},
	}
}

func TestPause_PausesJob(t *testing.T) {
	r := newWithClock(pauseJobs(), fixedNow)
	h := r.PauseHandler()

	req := httptest.NewRequest(http.MethodPost, "/pause?action=pause&job=backup", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.entries["backup"].Paused {
		t.Error("expected backup to be paused")
	}
}

func TestPause_ResumesJob(t *testing.T) {
	r := newWithClock(pauseJobs(), fixedNow)
	r.entries["backup"] = Entry{Job: r.entries["backup"].Job, LastSeen: time.Time{}, Paused: true}
	h := r.PauseHandler()

	req := httptest.NewRequest(http.MethodPost, "/pause?action=resume&job=backup", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.entries["backup"].Paused {
		t.Error("expected backup to be resumed")
	}
}

func TestPause_ListPaused(t *testing.T) {
	r := newWithClock(pauseJobs(), fixedNow)
	r.entries["report"] = Entry{Job: r.entries["report"].Job, LastSeen: time.Time{}, Paused: true}
	h := r.PauseHandler()

	req := httptest.NewRequest(http.MethodGet, "/pause", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string][]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp["paused"]) != 1 || resp["paused"][0] != "report" {
		t.Errorf("unexpected paused list: %v", resp["paused"])
	}
}

func TestPause_UnknownJobReturns404(t *testing.T) {
	r := newWithClock(pauseJobs(), fixedNow)
	h := r.PauseHandler()

	req := httptest.NewRequest(http.MethodPost, "/pause?action=pause&job=ghost", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPause_MissingJobParamReturns400(t *testing.T) {
	r := newWithClock(pauseJobs(), fixedNow)
	h := r.PauseHandler()

	req := httptest.NewRequest(http.MethodPost, "/pause?action=pause", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
