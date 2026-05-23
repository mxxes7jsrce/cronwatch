// Package jobregistry maintains a runtime registry of known cron jobs,
// tracking their schedule metadata and last-seen timestamps.
package jobregistry

import (
	"sync"
	"time"

	"github.com/cronwatch/internal/config"
)

// Entry holds runtime metadata for a single registered job.
type Entry struct {
	Name         string
	Schedule     string
	ExpectedEvery time.Duration
	LastSeen     time.Time
	RegisteredAt time.Time
}

// Registry is a thread-safe store of job entries.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	now     func() time.Time
}

// New creates a Registry pre-populated from the provided config jobs.
func New(jobs []config.Job) *Registry {
	return newWithClock(jobs, time.Now)
}

func newWithClock(jobs []config.Job, now func() time.Time) *Registry {
	r := &Registry{
		entries: make(map[string]*Entry, len(jobs)),
		now:     now,
	}
	for _, j := range jobs {
		r.entries[j.Name] = &Entry{
			Name:          j.Name,
			Schedule:      j.Schedule,
			ExpectedEvery: j.MaxAge,
			RegisteredAt:  now(),
		}
	}
	return r
}

// Touch records the current time as the last-seen timestamp for a job.
// If the job is not registered, Touch is a no-op and returns false.
func (r *Registry) Touch(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[name]
	if !ok {
		return false
	}
	e.LastSeen = r.now()
	return true
}

// Get returns a copy of the Entry for name, and whether it was found.
func (r *Registry) Get(name string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[name]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all registered entries.
func (r *Registry) All() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Entry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, *e)
	}
	return out
}

// Stale returns entries whose LastSeen is non-zero and older than ExpectedEvery.
func (r *Registry) Stale() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := r.now()
	var out []Entry
	for _, e := range r.entries {
		if e.LastSeen.IsZero() {
			continue
		}
		if now.Sub(e.LastSeen) > e.ExpectedEvery {
			out = append(out, *e)
		}
	}
	return out
}
