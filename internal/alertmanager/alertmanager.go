// Package alertmanager coordinates the evaluation of watcher results
// and dispatches notifications when a cron job is missed or has failed.
// It acts as the glue between the watcher, state store, and notifier.
package alertmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourorg/cronwatch/internal/config"
	"github.com/yourorg/cronwatch/internal/notify"
	"github.com/yourorg/cronwatch/internal/state"
	"github.com/yourorg/cronwatch/internal/watcher"
)

// Notifier is the interface used to send alert messages.
type Notifier interface {
	Send(ctx context.Context, msg string) error
}

// AlertManager evaluates watcher results and triggers alerts as needed.
type AlertManager struct {
	cfg      *config.Config
	store    *state.Store
	notifier Notifier
	watcher  *watcher.Watcher
}

// New creates a new AlertManager wired with the provided dependencies.
func New(cfg *config.Config, store *state.Store, n Notifier, w *watcher.Watcher) *AlertManager {
	return &AlertManager{
		cfg:      cfg,
		store:    store,
		notifier: n,
		watcher:  w,
	}
}

// Evaluate checks all configured jobs, updates state, and sends alerts
// for any job that is missed or has failed. It is intended to be called
// on each scheduler tick.
func (am *AlertManager) Evaluate(ctx context.Context) {
	for _, job := range am.cfg.Jobs {
		result, err := am.watcher.Check(ctx, job)
		if err != nil {
			log.Printf("[alertmanager] error checking job %q: %v", job.Name, err)
			continue
		}

		am.handleResult(ctx, job, result)
	}
}

// handleResult updates state and dispatches an alert based on the watcher result.
func (am *AlertManager) handleResult(ctx context.Context, job config.Job, result watcher.Result) {
	switch {
	case result.Missed:
		state.RecordMissed(am.store, job.Name)
		msg := am.formatAlert(job.Name, "missed", "no log activity detected within the expected window")
		am.sendAlert(ctx, job.Name, msg)

	case result.Failed:
		state.RecordFailure(am.store, job.Name)
		msg := am.formatAlert(job.Name, "failed", result.Reason)
		am.sendAlert(ctx, job.Name, msg)

	default:
		state.RecordSuccess(am.store, job.Name)
		log.Printf("[alertmanager] job %q is healthy", job.Name)
	}
}

// sendAlert dispatches a notification message, logging any delivery error.
func (am *AlertManager) sendAlert(ctx context.Context, jobName, msg string) {
	if err := am.notifier.Send(ctx, msg); err != nil {
		log.Printf("[alertmanager] failed to send alert for job %q: %v", jobName, err)
		return
	}
	log.Printf("[alertmanager] alert sent for job %q", jobName)
}

// formatAlert builds a human-readable alert string for a job event.
func (am *AlertManager) formatAlert(jobName, status, reason string) string {
	return fmt.Sprintf(
		"[cronwatch] job %q %s at %s — %s",
		jobName,
		status,
		time.Now().UTC().Format(time.RFC3339),
		reason,
	)
}

// NewFromConfig constructs an AlertManager using a webhook notifier derived
// from the provided configuration.
func NewFromConfig(cfg *config.Config, store *state.Store, w *watcher.Watcher) (*AlertManager, error) {
	n, err := notify.NewWebhookNotifier(cfg.Alerting.WebhookURL)
	if err != nil {
		return nil, fmt.Errorf("alertmanager: failed to create notifier: %w", err)
	}
	return New(cfg, store, n, w), nil
}
