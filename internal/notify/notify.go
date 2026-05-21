// Package notify provides alerting functionality for cronwatch.
// It supports sending notifications when cron jobs are missed or fail.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AlertType represents the kind of alert being sent.
type AlertType string

const (
	AlertMissed  AlertType = "missed"
	AlertFailed  AlertType = "failed"
	AlertTimeout AlertType = "timeout"
)

// Alert holds the details of a notification event.
type Alert struct {
	JobName   string    `json:"job_name"`
	AlertType AlertType `json:"alert_type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Notifier is the interface that wraps the Send method.
type Notifier interface {
	Send(alert Alert) error
}

// WebhookNotifier sends alerts to an HTTP webhook endpoint.
type WebhookNotifier struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with the given URL.
// A default timeout of 10 seconds is applied if timeout is zero.
func NewWebhookNotifier(url string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		URL:     url,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Send marshals the alert to JSON and POSTs it to the webhook URL.
func (w *WebhookNotifier) Send(alert Alert) error {
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now().UTC()
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("notify: marshal alert: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("notify: post to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
