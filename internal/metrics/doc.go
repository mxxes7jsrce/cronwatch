// Package metrics provides a lightweight, thread-safe counter registry for
// cronwatch operational telemetry.
//
// Counters are incremented by name throughout the daemon (monitor checks,
// alert dispatches, dedup drops, etc.) and exposed as a JSON snapshot via
// an HTTP handler suitable for mounting on the existing health-check server.
//
// Usage:
//
//	reg := metrics.New()
//	reg.Inc("checks_run")
//	http.Handle("/metrics", reg.Handler())
//
// Supported counter names:
//
//	"checks_run"    – number of watcher check cycles completed
//	"alerts_sent"   – number of webhook notifications dispatched
//	"missed_jobs"   – number of missed-job events detected
//	"failed_jobs"   – number of failed-job events detected
//	"dedup_dropped" – number of alerts suppressed by deduplication
package metrics
