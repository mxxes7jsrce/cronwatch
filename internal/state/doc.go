// Package state provides persistent storage and lifecycle tracking for
// monitored cron jobs.
//
// A Store holds JobState records on disk (JSON) and is safe for concurrent
// access. Helper functions in updater.go record success, failure, and missed
// events, and expose a simple predicate for determining whether a job has
// exceeded its expected execution window.
//
// Typical usage:
//
//	store, err := state.NewStore("/var/lib/cronwatch/state.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// When a job heartbeat arrives:
//	_ = state.RecordSuccess(store, "daily-backup")
//
//	// During a periodic check:
//	js := store.Get("daily-backup")
//	if state.IsMissed(js, 24*time.Hour, 10*time.Minute) {
//		// send alert
//	}
package state
