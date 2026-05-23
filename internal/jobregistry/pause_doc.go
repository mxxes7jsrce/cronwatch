// Package jobregistry provides job lifecycle tracking for cronwatch.
//
// # Pause / Resume
//
// The PauseHandler exposes three operations over HTTP:
//
//	GET  /pause                        — returns a JSON list of all currently paused jobs
//	POST /pause?action=pause&job=NAME  — marks a job as paused
//	POST /pause?action=resume&job=NAME — clears the paused flag on a job
//
// While a job is paused the staleness and expiry checkers skip it, so
// no missed-run alerts are generated for that job.
//
// Paused state is held in memory only; it resets when the daemon restarts.
package jobregistry
