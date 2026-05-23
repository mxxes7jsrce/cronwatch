package jobregistry

import (
	"net/http"
	"sort"
	"time"
)

// StalenessReport summarises jobs that have not checked in within their
// expected interval.
type StalenessReport struct {
	GeneratedAt time.Time      `json:"generated_at"`
	StaleJobs   []StaleJobInfo `json:"stale_jobs"`
	TotalStale  int            `json:"total_stale"`
}

// StaleJobInfo holds staleness details for a single job.
type StaleJobInfo struct {
	Name        string        `json:"name"`
	LastSeen    time.Time     `json:"last_seen"`
	SilentFor   time.Duration `json:"silent_for_ns"`
	MaxInterval time.Duration `json:"max_interval_ns"`
}

// buildStalenessReport inspects all registry entries and returns a report of
// jobs whose last-seen timestamp exceeds their configured max interval.
func (r *Registry) buildStalenessReport() StalenessReport {
	now := r.clock()

	r.mu.RLock()
	defer r.mu.RUnlock()

	var stale []StaleJobInfo
	for name, entry := range r.entries {
		if entry.LastSeen.IsZero() {
			continue
		}
		silentFor := now.Sub(entry.LastSeen)
		if silentFor > entry.MaxInterval {
			stale = append(stale, StaleJobInfo{
				Name:        name,
				LastSeen:    entry.LastSeen,
				SilentFor:   silentFor,
				MaxInterval: entry.MaxInterval,
			})
		}
	}

	sort.Slice(stale, func(i, j int) bool {
		return stale[i].Name < stale[j].Name
	})

	return StalenessReport{
		GeneratedAt: now,
		StaleJobs:   stale,
		TotalStale:  len(stale),
	}
}

// StalenessHandler returns an HTTP handler that exposes the staleness report
// as JSON. It is intended to be mounted at GET /jobs/stale.
func (r *Registry) StalenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		report := r.buildStalenessReport()
		writeJSON(w, report)
	})
}
