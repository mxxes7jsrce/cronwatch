package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/healthcheck"
	"github.com/example/cronwatch/internal/state"
)

func newTestStore(t *testing.T) *state.Store {
	t.Helper()
	p := filepath.Join(t.TempDir(), "state.json")
	s, err := state.NewStore(p)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestHealthHandler_EmptyStore(t *testing.T) {
	store := newTestStore(t)
	srv := healthcheck.New(":0", store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp healthcheck.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %q", resp.Status)
	}
	if resp.JobCount != 0 {
		t.Errorf("expected 0 jobs, got %d", resp.JobCount)
	}
}

func TestHealthHandler_CountsMissedAndFailed(t *testing.T) {
	store := newTestStore(t)

	store.Set("job-a", state.Entry{FailureCount: 2})
	store.Set("job-b", state.Entry{MissedCount: 1})
	store.Set("job-c", state.Entry{LastSuccess: time.Now()})

	srv := healthcheck.New(":0", store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)

	var resp healthcheck.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.JobCount != 3 {
		t.Errorf("expected 3 jobs, got %d", resp.JobCount)
	}
	if resp.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", resp.Failed)
	}
	if resp.Missed != 1 {
		t.Errorf("expected 1 missed, got %d", resp.Missed)
	}
}

func TestHealthHandler_ContentTypeJSON(t *testing.T) {
	store := newTestStore(t)
	srv := healthcheck.New(":0", store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
