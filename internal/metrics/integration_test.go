package metrics_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/internal/metrics"
)

// TestMetrics_EndToEndHTTP verifies the full path: increment counters then
// retrieve them through the HTTP handler, simulating real daemon usage.
func TestMetrics_EndToEndHTTP(t *testing.T) {
	reg := metrics.New()

	events := []struct {
		name  string
		count int
	}{
		{"checks_run", 5},
		{"alerts_sent", 2},
		{"missed_jobs", 1},
		{"failed_jobs", 3},
		{"dedup_dropped", 4},
	}

	for _, e := range events {
		for i := 0; i < e.count; i++ {
			reg.Inc(e.name)
		}
	}

	srv := httptest.NewServer(reg.Handler())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}

	var result map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, e := range events {
		if got := result[e.name]; got != int64(e.count) {
			t.Errorf("%s: want %d, got %d", e.name, e.count, got)
		}
	}
}
