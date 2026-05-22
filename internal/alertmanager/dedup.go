package alertmanager

import (
	"sync"
	"time"
)

// dedupKey uniquely identifies an alert by job name and alert kind.
type dedupKey struct {
	JobName string
	Kind    string
}

// dedupStore tracks the last time an alert was sent for each key,
// enabling cooldown-based deduplication.
type dedupStore struct {
	mu      sync.Mutex
	records map[dedupKey]time.Time
	clock   func() time.Time
}

func newDedupStore(clock func() time.Time) *dedupStore {
	if clock == nil {
		clock = time.Now
	}
	return &dedupStore{
		records: make(map[dedupKey]time.Time),
		clock:   clock,
	}
}

// allow returns true if the alert should be sent (i.e. not within cooldown).
// If allowed, it records the current time for the key.
func (d *dedupStore) allow(jobName, kind string, cooldown time.Duration) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := dedupKey{JobName: jobName, Kind: kind}
	now := d.clock()

	if last, ok := d.records[key]; ok {
		if now.Sub(last) < cooldown {
			return false
		}
	}

	d.records[key] = now
	return true
}

// reset clears the dedup record for a specific job and kind.
func (d *dedupStore) reset(jobName, kind string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.records, dedupKey{JobName: jobName, Kind: kind})
}
