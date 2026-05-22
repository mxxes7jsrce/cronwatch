package alertmanager

import (
	"sync"
	"time"
)

type dedupKey struct {
	job  string
	kind string
}

type dedupStore struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[dedupKey]time.Time
	clock    func() time.Time
}

func newDedupStore(cooldown time.Duration) *dedupStore {
	return &dedupStore{
		cooldown: cooldown,
		last:     make(map[dedupKey]time.Time),
		clock:    time.Now,
	}
}

// allow returns true if an alert for (job, kind) should be sent.
// It records the current time as the last alert time when returning true.
func (d *dedupStore) allow(job, kind string) bool {
	if d.cooldown == 0 {
		return true
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	k := dedupKey{job: job, kind: kind}
	now := d.clock()
	if t, ok := d.last[k]; ok && now.Sub(t) < d.cooldown {
		return false
	}
	d.last[k] = now
	return true
}

// reset removes the dedup entry so the next alert is not suppressed.
func (d *dedupStore) reset(job, kind string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.last, dedupKey{job: job, kind: kind})
}
