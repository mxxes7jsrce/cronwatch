package jobregistry

import (
	"time"
)

// ExpiryChecker scans the registry for jobs that have not been touched
// within their configured timeout and returns their names.
type ExpiryChecker struct {
	reg *Registry
}

// NewExpiryChecker returns an ExpiryChecker backed by the given Registry.
func NewExpiryChecker(r *Registry) *ExpiryChecker {
	return &ExpiryChecker{reg: r}
}

// Expired returns the names of all jobs whose last-seen time is older than
// their individual timeout as of the given reference time t.
func (e *ExpiryChecker) Expired(t time.Time) []string {
	e.reg.mu.RLock()
	defer e.reg.mu.RUnlock()

	var expired []string
	for name, entry := range e.reg.entries {
		deadline := entry.LastSeen.Add(entry.Timeout)
		if t.After(deadline) {
			expired = append(expired, name)
		}
	}
	return expired
}

// NextExpiry returns the earliest time at which any tracked job will expire,
// and a boolean indicating whether any jobs are registered.
func (e *ExpiryChecker) NextExpiry() (time.Time, bool) {
	e.reg.mu.RLock()
	defer e.reg.mu.RUnlock()

	var earliest time.Time
	found := false
	for _, entry := range e.reg.entries {
		candidate := entry.LastSeen.Add(entry.Timeout)
		if !found || candidate.Before(earliest) {
			earliest = candidate
			found = true
		}
	}
	return earliest, found
}
