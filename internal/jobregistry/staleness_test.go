package jobregistry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

var stalenessJobs = []config.Job{
	{Name: "alpha", Schedule: "@every 1m", LogPath: "/tmp/alpha.log"},
	{Name: "beta", Schedule: "@every 1m", LogPath: "/tmp/beta.log"},
	{Name: "gamma", Schedule: "@every 1m", LogPath: "/tmp/gamma.log"},
}

func TestStaleness_NoneStale(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(stalenessJobs, func() time.Time { return now })

	// Touch all jobs just now — none should be stale.
	for _, j := range stalenessJobs {
		reg.Touch(j.Name)
	}

	report := reg.buildStalenessReport()
	if report.TotalStale != 0 {
		t.Fatalf("expected 0 stale jobs, got %d", report.TotalStale)
	}
	if len(report.StaleJobs) != 0 {
		t.Fatalf("expected empty stale list")
	}
}

func TestStaleness_OneStale(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return base }
	reg := newWithClock(stalenessJobs, clock)

	// Touch alpha and gamma recently, leave beta un-touched (LastSeen zero).
	reg.Touch("alpha")
	reg.Touch("gamma")

	// Advance clock past max interval for beta by manually setting its entry.
	reg.mu.Lock()
	e := reg.entries["beta"]
	e.LastSeen = base.Add(-10 * time.Minute)
	reg.entries["beta"] = e
	reg.mu.Unlock()

	report := reg.buildStalenessReport()
	if report.TotalStale != 1 {
		t.Fatalf("expected 1 stale job, got %d", report.TotalStale)
	}
	if report.StaleJobs[0].Name != "beta" {
		t.Errorf("expected stale job 'beta', got %q", report.StaleJobs[0].Name)
	}
	if report.StaleJobs[0].SilentFor <= report.StaleJobs[0].MaxInterval {
		t.Errorf("SilentFor should exceed MaxInterval")
	}
}

func TestStaleness_HandlerReturnsJSON(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(stalenessJobs, func() time.Time { return now })

	req := httptest.NewRequest(http.MethodGet, "/jobs/stale", nil)
	rec := httptest.NewRecorder()
	reg.StalenessHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}

	var report StalenessReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestStaleness_HandlerMethodNotAllowed(t *testing.T) {
	reg := newWithClock(stalenessJobs, time.Now)

	req := httptest.NewRequest(http.MethodPost, "/jobs/stale", nil)
	rec := httptest.NewRecorder()
	reg.StalenessHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
