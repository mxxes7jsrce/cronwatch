package watcher

import "time"

// Result holds the outcome of a single watcher check for a cron job.
type Result struct {
	// JobName is the name of the cron job that was checked.
	JobName string

	// Success indicates whether the job is considered to have run successfully.
	Success bool

	// Output is the last N lines (or full content) read from the log file.
	Output string

	// MatchedKeyword is the failure keyword found in the output, if any.
	// Empty string means no failure keyword was detected.
	MatchedKeyword string

	// CheckedAt is the time at which the check was performed.
	CheckedAt time.Time
}
