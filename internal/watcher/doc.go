// Package watcher provides functionality for monitoring log files and
// command output to detect cron job execution results.
//
// The watcher reads from a configured source (log file or command stdout/stderr)
// and determines whether a cron job succeeded or failed based on exit codes
// and configurable failure keywords found in the output.
//
// Basic usage:
//
//	w, err := watcher.New(cfg, jobName)
//	if err != nil {
//		log.Fatal(err)
//	}
//	result, err := w.Check(ctx)
package watcher
