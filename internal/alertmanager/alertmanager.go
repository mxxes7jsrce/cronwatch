// Package alertmanager coordinates alert evaluation and delivery.
package alertmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/notify"
	"github.com/cronwatch/internal/watcher"
)

// Notifier is the interface for sending alert messages.
type Notifier interface {
	Send(ctx context.Context, message string) error
}

// AlertManager evaluates watcher results and dispatches alerts.
type AlertManager struct {
	notifier Notifier
	dedup    *dedupStore
	clock    func() time.Time
}

// New creates an AlertManager using the provided config and notifier.
func New(cfg config.Config, n Notifier) *AlertManager {
	cooldown := time.Duration(cfg.AlertCooldownMinutes) * time.Minute
	return &AlertManager{
		notifier: n,
		dedup:    newDedupStore(cooldown),
		clock:    time.Now,
	}
}

// newWithClock creates an AlertManager with an injectable clock (for tests).
func newWithClock(cfg config.Config, n Notifier, clock func() time.Time) *AlertManager {
	cooldown := time.Duration(cfg.AlertCooldownMinutes) * time.Minute
	return &AlertManager{
		notifier: n,
		dedup:    newDedupStoreClock(cooldown, clock),
		clock:    clock,
	}
}

// Evaluate inspects a watcher.Result and sends an alert if warranted.
func (am *AlertManager) Evaluate(ctx context.Context, job config.Job, result watcher.Result) {
	if result.Healthy {
		return
	}

	kind := alertKind(result)
	if !am.dedup.allow(job.Name, kind, am.clock()) {
		log.Printf("alertmanager: suppressing duplicate alert for job=%s kind=%s", job.Name, kind)
		return
	}

	msg := formatMessage(job, result)
	if err := am.notifier.Send(ctx, msg); err != nil {
		log.Printf("alertmanager: failed to send alert for job=%s: %v", job.Name, err)
	}
}

func alertKind(r watcher.Result) string {
	if r.Missed {
		return "missed"
	}
	return "failure"
}

func formatMessage(job config.Job, result watcher.Result) string {
	if result.Missed {
		return fmt.Sprintf("[cronwatch] MISSED: job %q has not run since %s",
			job.Name, result.LastSeen.Format(time.RFC3339))
	}
	return fmt.Sprintf("[cronwatch] FAILURE: job %q matched failure keyword %q in log %s",
		job.Name, result.MatchedKeyword, job.LogPath)
}

// ensure compile-time check that notify.WebhookNotifier satisfies Notifier.
var _ Notifier = (*notify.WebhookNotifier)(nil)
