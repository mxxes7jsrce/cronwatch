package jobregistry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/jobregistry"
)

func TestHeartbeat_Integration_TouchUpdatesLastSeen(t *testing.T) {
	jobs := []config.Job{
		{Name: "sync", Schedule: "@hourly", LogPath: "/var/log/sync.log", Timeout: 30 * time.Minute},
	}

	r := jobregistry.New(jobs)
	server := httptest.NewServer(r.HeartbeatHandler())
	t.Cleanup(server.Close)

	// Manually set path value via a real ServeMux for integration
	mux := http.NewServeMux()
	mux.Handle("POST /heartbeat/{job}", r.HeartbeatHandler())
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/heartbeat/sync", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	entry, ok := r.Get("sync")
	if !ok {
		t.Fatal("expected entry for sync")
	}
	if entry.LastSeen.IsZero() {
		t.Error("expected LastSeen to be set after heartbeat")
	}
}

func TestHeartbeat_Integration_UnknownJob(t *testing.T) {
	jobs := []config.Job{
		{Name: "sync", Schedule: "@hourly", LogPath: "/var/log/sync.log", Timeout: 30 * time.Minute},
	}

	r := jobregistry.New(jobs)
	mux := http.NewServeMux()
	mux.Handle("POST /heartbeat/{job}", r.HeartbeatHandler())
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/heartbeat/nope", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
