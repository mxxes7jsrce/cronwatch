// Package notify provides alerting mechanisms for cronwatch.
//
// It defines the Notifier interface and concrete implementations
// for sending alerts when cron jobs are missed or fail.
//
// # Notifier Interface
//
// Any type that implements Notifier can be used to dispatch alerts:
//
//	type Notifier interface {
//		Send(ctx context.Context, alert Alert) error
//	}
//
// # WebhookNotifier
//
// WebhookNotifier sends JSON-encoded Alert payloads via HTTP POST to a
// configured endpoint. It supports a configurable timeout and includes
// a timestamp in every outbound payload.
//
// Example configuration:
//
//	notifier := notify.NewWebhookNotifier("https://hooks.example.com/alert", 10*time.Second)
//	err := notifier.Send(ctx, notify.Alert{
//		JobName: "backup",
//		Status:  "missed",
//		Message: "Job has not run within the expected window",
//	})
//
// # Alert Payload
//
// Alerts are serialised as JSON and include the job name, status,
// a human-readable message, and an RFC3339 timestamp added automatically
// by the notifier at send time.
package notify
