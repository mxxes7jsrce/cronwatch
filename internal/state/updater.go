package state

import (
	"time"
)

// RecordSuccess marks a job as successfully seen at the given time.
func RecordSuccess(store *Store, name string) error {
	existing := store.Get(name)
	js := mergeOrNew(existing, name)
	js.LastSeen = time.Now().UTC()
	js.LastStatus = "ok"
	js.FailCount = 0
	return store.Set(js)
}

// RecordFailure increments the failure counter for a job.
func RecordFailure(store *Store, name string) error {
	existing := store.Get(name)
	js := mergeOrNew(existing, name)
	js.LastSeen = time.Now().UTC()
	js.LastStatus = "failed"
	js.FailCount++
	return store.Set(js)
}

// RecordMissed increments the missed counter for a job without updating LastSeen.
func RecordMissed(store *Store, name string) error {
	existing := store.Get(name)
	js := mergeOrNew(existing, name)
	js.LastStatus = "missed"
	js.MissedCount++
	return store.Set(js)
}

// IsMissed returns true when the job's last seen time is older than its interval.
func IsMissed(js *JobState, interval time.Duration, grace time.Duration) bool {
	if js == nil || js.LastSeen.IsZero() {
		return false
	}
	deadline := js.LastSeen.Add(interval).Add(grace)
	return time.Now().UTC().After(deadline)
}

func mergeOrNew(existing *JobState, name string) *JobState {
	if existing != nil {
		copy := *existing
		return &copy
	}
	return &JobState{Name: name}
}
