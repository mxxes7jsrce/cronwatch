package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/notify"
	"github.com/cronwatch/internal/state"
)

func buildDeps(t *testing.T, webhookURL string) (*config.Config, *state.Store, *notify.WebhookNotifier) {
	t.Helper()
	cfg := &config.Config{
		CheckInterval: 50 * time.Millisecond,
		Jobs: []config.Job{
			{Name: "backup", GracePeriod: 1 * time.Hour, MaxFailures: 3},
		},
	}
	store, err := state.NewStore(t.TempDir() + "/state.json")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	notifier, _ := notify.NewWebhookNotifier(webhookURL, 2*time.Second)
	return cfg, store, notifier
}

func TestMonitor_AlertsOnMissedJob(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["kind"] == "missed" {
			calls.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg, store, notifier := buildDeps(t, ts.URL)
	// Job has never run — IsMissed should fire immediately (zero time, large grace).
	// Force grace period to zero so the check triggers right away.
	cfg.Jobs[0].GracePeriod = 0

	m := New(cfg, store, notifier)
	done := make(chan struct{})
	go m.Run(done)

	time.Sleep(200 * time.Millisecond)
	close(done)

	if calls.Load() == 0 {
		t.Error("expected at least one missed alert, got none")
	}
}

func TestMonitor_NoAlertWhenJobHealthy(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg, store, notifier := buildDeps(t, ts.URL)
	// Record a recent success so the job is not considered missed.
	state.RecordSuccess(store, "backup")

	m := New(cfg, store, notifier)
	done := make(chan struct{})
	go m.Run(done)

	time.Sleep(200 * time.Millisecond)
	close(done)

	if calls.Load() != 0 {
		t.Errorf("expected no alerts for healthy job, got %d", calls.Load())
	}
}
