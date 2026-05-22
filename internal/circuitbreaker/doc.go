// Package circuitbreaker provides a thread-safe circuit breaker for use with
// outbound notification calls in cronwatch.
//
// The circuit breaker transitions through three states:
//
//   - Closed: normal operation; all calls are allowed through.
//   - Open: the notifier has exceeded the failure threshold; calls are blocked
//     until the reset timeout elapses.
//   - Half-Open: a probe call is allowed to test whether the downstream service
//     has recovered; a success closes the circuit, a failure reopens it.
//
// Usage:
//
//	breaker := circuitbreaker.New(5, 30*time.Second)
//
//	if !breaker.Allow() {
//		return circuitbreaker.ErrCircuitOpen
//	}
//	if err := notifier.Send(ctx, msg); err != nil {
//		breaker.RecordFailure()
//		return err
//	}
//	breaker.RecordSuccess()
package circuitbreaker
