// Package retrier implements a generic retry mechanism with configurable
// exponential backoff, suitable for wrapping any fallible operation such
// as webhook delivery or external API calls.
//
// # Usage
//
//	cfg := retrier.DefaultConfig()
//	r := retrier.New(cfg)
//	err := r.Do(ctx, func(attempt int) error {
//	    return sendAlert()
//	})
//	if errors.Is(err, retrier.ErrMaxAttempts) {
//	    log.Println("alert delivery failed after all retries")
//	}
//
// # Backoff
//
// The delay between attempts starts at BaseDelay and is multiplied by
// Multiplier after each failure, capped at MaxDelay. The retry loop
// respects context cancellation between attempts.
package retrier
