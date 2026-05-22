// Package healthcheck exposes a lightweight HTTP liveness endpoint for the
// cronwatch daemon.
//
// Usage:
//
//	store := state.NewStore(path)
//	srv := healthcheck.New(":9090", store)
//	go srv.ListenAndServe()
//
// The /healthz endpoint returns a JSON payload containing:
//   - status: always "ok" when the daemon is running
//   - uptime: human-readable duration since the daemon started
//   - checked_at: UTC timestamp of the request
//   - job_count: total number of tracked jobs
//   - missed_jobs: number of jobs that have at least one missed run
//   - failed_jobs: number of jobs that have at least one recorded failure
package healthcheck
