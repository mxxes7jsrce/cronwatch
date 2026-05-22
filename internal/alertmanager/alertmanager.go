// Package alertmanager coordinates alert dispatch with deduplication.
package alertmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cronwatch/internal/watcher"
)

// Notifier is the interface for sending alert messages.
type Notifier interface {
	Send(ctx context.Context, message string) error
}

// Manager handles alert routing and deduplication.
type Manager struct {
	notifier Notifier
	dedup    *dedupStore
}

// New creates a new Manager with the given notifier and cooldown duration.
// A cooldown of 0 disables deduplication.
func New(n Notifier, cooldown time.Duration) *Manager {
	return &Manager{
		notifier: n,
		dedup:    newDedupStore(cooldown),
	}
}

// Handle processes a watcher result and sends an alert if warranted.
func (m *Manager) Handle(ctx context.Context, res watcher.Result) {
	if !res.Failed && !res.Missed {
		return
	}

	kind := alertKind(res)
	if !m.dedup.allow(res.Job.Name, kind) {
		log.Printf("[alertmanager] suppressed duplicate alert job=%s kind=%s", res.Job.Name, kind)
		return
	}

	msg := formatMessage(res)
	if err := m.notifier.Send(ctx, msg); err != nil {
		log.Printf("[alertmanager] failed to send alert for job=%s: %v", res.Job.Name, err)
		// roll back dedup entry so the next attempt is not suppressed
		m.dedup.reset(res.Job.Name, kind)
	}
}

// HandleBatch processes multiple results.
func (m *Manager) HandleBatch(ctx context.Context, results []watcher.Result) {
	for _, r := range results {
		m.Handle(ctx, r)
	}
}

func alertKind(res watcher.Result) string {
	if res.Missed {
		return "missed"
	}
	return "failed"
}

func formatMessage(res watcher.Result) string {
	kind := alertKind(res)
	return fmt.Sprintf(
		"[cronwatch] job %q %s at %s — %s",
		res.Job.Name,
		kind,
		time.Now().UTC().Format(time.RFC3339),
		res.Reason,
	)
}
