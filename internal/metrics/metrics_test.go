package metrics_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/internal/metrics"
)

func TestRegistry_IncAndSnapshot(t *testing.T) {
	r := metrics.New()

	r.Inc("checks_run")
	r.Inc("checks_run")
	r.Inc("alerts_sent")
	r.Inc("missed_jobs")
	r.Inc("failed_jobs")
	r.Inc("dedup_dropped")

	snap := r.Snapshot()

	if got := snap["checks_run"]; got != 2 {
		t.Errorf("checks_run: want 2, got %d", got)
	}
	if got := snap["alerts_sent"]; got != 1 {
		t.Errorf("alerts_sent: want 1, got %d", got)
	}
	if got := snap["missed_jobs"]; got != 1 {
		t.Errorf("missed_jobs: want 1, got %d", got)
	}
	if got := snap["failed_jobs"]; got != 1 {
		t.Errorf("failed_jobs: want 1, got %d", got)
	}
	if got := snap["dedup_dropped"]; got != 1 {
		t.Errorf("dedup_dropped: want 1, got %d", got)
	}
}

func TestRegistry_UnknownCounterIgnored(t *testing.T) {
	r := metrics.New()
	r.Inc("nonexistent_counter")
	snap := r.Snapshot()
	for k, v := range snap {
		if v != 0 {
			t.Errorf("expected zero for %s after unknown inc, got %d", k, v)
		}
	}
}

func TestRegistry_HandlerReturnsJSON(t *testing.T) {
	r := metrics.New()
	r.Inc("checks_run")
	r.Inc("failed_jobs")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	r.Handler()(w, req)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %s", ct)
	}

	var result map[string]int64
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["checks_run"] != 1 {
		t.Errorf("checks_run: want 1, got %d", result["checks_run"])
	}
	if result["failed_jobs"] != 1 {
		t.Errorf("failed_jobs: want 1, got %d", result["failed_jobs"])
	}
}

func TestRegistry_ConcurrentInc(t *testing.T) {
	r := metrics.New()
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			r.Inc("checks_run")
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	if got := r.Snapshot()["checks_run"]; got != 100 {
		t.Errorf("concurrent inc: want 100, got %d", got)
	}
}
