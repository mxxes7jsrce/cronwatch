// Package alertmanager coordinates alert dispatch for cronwatch.
//
// It receives check results from the watcher and monitor packages,
// deduplicates alerts to prevent notification storms, and delegates
// delivery to the configured notifier (e.g. webhook).
//
// # Deduplication
//
// Each job alert is suppressed if an identical alert was sent within
// the configured cooldown window. This prevents repeated pages for a
// job that remains in a failed or missed state across multiple check
// intervals.
//
// # Usage
//
//	am := alertmanager.New(notifier, cooldown)
//	am.Send(ctx, alert)
package alertmanager
