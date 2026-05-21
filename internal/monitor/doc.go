// Package monitor implements the core monitoring loop for cronwatch.
//
// A Monitor is constructed with a parsed Config, a state Store, and a
// WebhookNotifier.  Calling Run blocks and ticks at the configured
// CheckInterval; on each tick it evaluates every job defined in the config:
//
//   - If a job has not run within its GracePeriod it is considered "missed"
//     and an alert is dispatched via the notifier.
//
//   - If a job's consecutive failure count meets or exceeds MaxFailures
//     a "failed" alert is dispatched.
//
// Run returns only when the done channel is closed, making it easy to
// integrate with os/signal handling or context cancellation in main.
package monitor
