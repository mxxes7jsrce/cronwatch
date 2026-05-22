// Package metrics provides lightweight in-process counters for cronwatch
// operational telemetry, exposed via a /metrics HTTP endpoint.
package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
)

// Counters holds cumulative event counts for the daemon lifetime.
type Counters struct {
	ChecksRun     atomic.Int64 `json:"checks_run"`
	Alertssent    atomic.Int64 `json:"alerts_sent"`
	MissedJobs    atomic.Int64 `json:"missed_jobs"`
	FailedJobs    atomic.Int64 `json:"failed_jobs"`
	DedupDropped  atomic.Int64 `json:"dedup_dropped"`
}

// Registry holds a single shared Counters instance.
type Registry struct {
	mu       sync.RWMutex
	counters *Counters
}

// New returns a new Registry with zeroed counters.
func New() *Registry {
	return &Registry{counters: &Counters{}}
}

// Inc increments the named counter by 1. Unknown names are silently ignored.
func (r *Registry) Inc(name string) {
	switch name {
	case "checks_run":
		r.counters.ChecksRun.Add(1)
	case "alerts_sent":
		r.counters.Alertsent.Add(1)
	case "missed_jobs":
		r.counters.MissedJobs.Add(1)
	case "failed_jobs":
		r.counters.FailedJobs.Add(1)
	case "dedup_dropped":
		r.counters.DedupDropped.Add(1)
	}
}

// Snapshot returns a point-in-time copy of counter values.
func (r *Registry) Snapshot() map[string]int64 {
	return map[string]int64{
		"checks_run":    r.counters.ChecksRun.Load(),
		"alerts_sent":   r.counters.Alertsent.Load(),
		"missed_jobs":   r.counters.MissedJobs.Load(),
		"failed_jobs":   r.counters.FailedJobs.Load(),
		"dedup_dropped": r.counters.DedupDropped.Load(),
	}
}

// Handler returns an http.HandlerFunc that serves counter snapshots as JSON.
func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		snap := r.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snap)
	}
}
