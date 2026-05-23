// Package jobregistry tracks the last-seen heartbeat for each configured cron
// job and exposes helpers for detecting stale (expired) jobs.
//
// # ExpiryChecker
//
// ExpiryChecker wraps a Registry and provides two methods:
//
//   - Expired(t) – returns the names of all jobs whose last-seen timestamp
//     plus their configured timeout is before t.
//
//   - NextExpiry() – returns the earliest absolute time at which any job
//     will transition from healthy to expired, useful for scheduling the
//     next monitor wake-up with minimal busy-waiting.
//
// Both methods acquire only a read-lock on the registry so they are safe to
// call concurrently with Touch and Get.
package jobregistry
