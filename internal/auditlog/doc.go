// Package auditlog provides a structured, append-only audit trail for
// cronwatch daemon events.
//
// Each event is written as a newline-delimited JSON record containing a
// timestamp, event kind, job name, and optional message.  The log file is
// safe for concurrent use and is designed to be tailed or ingested by
// external log-aggregation tooling.
//
// Supported event kinds:
//
//	EventAlert   – an alert notification was dispatched
//	EventMissed  – a cron job was detected as missed
//	EventFailure – a cron job log indicated a failure keyword
//	EventSuccess – a cron job completed successfully
package auditlog
