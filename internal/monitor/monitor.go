package monitor

import (
	"log"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/notify"
	"github.com/cronwatch/internal/state"
)

// Monitor checks cron job states against their expected schedules
// and fires alerts when jobs are missed or have failed too many times.
type Monitor struct {
	cfg      *config.Config
	store    *state.Store
	notifier *notify.WebhookNotifier
}

// New creates a Monitor with the given dependencies.
func New(cfg *config.Config, store *state.Store, notifier *notify.WebhookNotifier) *Monitor {
	return &Monitor{
		cfg:      cfg,
		store:    store,
		notifier: notifier,
	}
}

// Run starts the monitoring loop, ticking at the configured check interval.
// It blocks until the provided done channel is closed.
func (m *Monitor) Run(done <-chan struct{}) {
	ticker := time.NewTicker(m.cfg.CheckInterval)
	defer ticker.Stop()

	log.Printf("monitor: starting, check interval=%s", m.cfg.CheckInterval)

	for {
		select {
		case <-ticker.C:
			m.check()
		case <-done:
			log.Println("monitor: shutting down")
			return
		}
	}
}

// check evaluates every configured job and sends alerts where necessary.
func (m *Monitor) check() {
	now := time.Now()
	for _, job := range m.cfg.Jobs {
		entry, ok := m.store.Get(job.Name)
		if !ok {
			// Never seen — treat as missed if the grace window has elapsed.
			if state.IsMissed(entry, job.GracePeriod, now) {
				m.alert(job.Name, "missed", "job has never run")
			}
			continue
		}

		if state.IsMissed(entry, job.GracePeriod, now) {
			m.alert(job.Name, "missed", "job did not run within grace period")
		}

		if job.MaxFailures > 0 && entry.ConsecutiveFailures >= job.MaxFailures {
			m.alert(job.Name, "failed",
				"consecutive failures exceeded threshold")
		}
	}
}

func (m *Monitor) alert(jobName, kind, reason string) {
	if err := m.notifier.Send(jobName, kind, reason); err != nil {
		log.Printf("monitor: alert send error job=%s kind=%s: %v", jobName, kind, err)
	}
}
