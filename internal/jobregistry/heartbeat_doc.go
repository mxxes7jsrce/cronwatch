// Package jobregistry provides the HeartbeatHandler, which exposes an HTTP
// endpoint that external cron wrappers or scripts can POST to after a
// successful job execution.
//
// # Endpoint
//
//	POST /heartbeat/{job}
//
// On success the handler calls Registry.Touch to update the job's LastSeen
// timestamp, which is used by ExpiryChecker to determine whether a job has
// missed its expected execution window.
//
// # Response
//
// HTTP 200 with a JSON body: {"status":"ok","job":"<name>","ts":"<rfc3339>"}
//
// Error codes:
//   - 400 – job name missing from path
//   - 404 – job name not registered
//   - 405 – non-POST method used
package jobregistry
