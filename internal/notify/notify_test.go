package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received notify.Alert

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json content-type, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewWebhookNotifier(server.URL, 5*time.Second)
	alert := notify.Alert{
		JobName:   "backup",
		AlertType: notify.AlertMissed,
		Message:   "job did not run within expected window",
		Timestamp: time.Now().UTC(),
	}

	if err := n.Send(alert); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if received.JobName != alert.JobName {
		t.Errorf("job name mismatch: got %q, want %q", received.JobName, alert.JobName)
	}
	if received.AlertType != alert.AlertType {
		t.Errorf("alert type mismatch: got %q, want %q", received.AlertType, alert.AlertType)
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := notify.NewWebhookNotifier(server.URL, 5*time.Second)
	err := n.Send(notify.Alert{JobName: "test", AlertType: notify.AlertFailed})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookNotifier_Send_SetsTimestamp(t *testing.T) {
	var received notify.Alert

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	n := notify.NewWebhookNotifier(server.URL, 0)
	alert := notify.Alert{JobName: "cleanup", AlertType: notify.AlertTimeout}
	// Timestamp intentionally left zero

	if err := n.Send(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically, got zero")
	}
}

func TestNewWebhookNotifier_DefaultTimeout(t *testing.T) {
	n := notify.NewWebhookNotifier("http://example.com", 0)
	if n.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", n.Timeout)
	}
}
