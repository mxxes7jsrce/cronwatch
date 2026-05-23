// Package jobregistry provides a thread-safe runtime registry of cron jobs
// loaded from configuration. It tracks when each job was last observed running
// and exposes helpers to identify stale (overdue) jobs.
//
// Typical usage:
//
//	reg := jobregistry.New(cfg.Jobs)
//
//	// When a job run is detected:
//	reg.Touch("backup-db")
//
//	// Periodically check for overdue jobs:
//	for _, e := range reg.Stale() {
//		log.Printf("job %s is overdue", e.Name)
//	}
package jobregistry
