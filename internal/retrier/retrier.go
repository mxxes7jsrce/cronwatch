// Package retrier provides a configurable retry mechanism with exponential
// backoff for use when sending alerts or making outbound HTTP calls.
package retrier

import (
	"context"
	"errors"
	"time"
)

// ErrMaxAttempts is returned when all retry attempts are exhausted.
var ErrMaxAttempts = errors.New("retrier: max attempts reached")

// Config holds the retry policy parameters.
type Config struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// BaseDelay is the initial backoff duration.
	BaseDelay time.Duration
	// MaxDelay caps the exponential backoff.
	MaxDelay time.Duration
	// Multiplier is applied to the delay after each failure.
	Multiplier float64
}

// DefaultConfig returns a sensible default retry policy.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
	}
}

// Retrier executes a function with retry logic.
type Retrier struct {
	cfg   Config
	sleep func(context.Context, time.Duration) error
}

// New creates a Retrier with the given Config.
func New(cfg Config) *Retrier {
	return &Retrier{
		cfg:   cfg,
		sleep: contextSleep,
	}
}

// Do calls fn up to MaxAttempts times. It stops early if ctx is cancelled.
// fn receives the current attempt number (1-based).
func (r *Retrier) Do(ctx context.Context, fn func(attempt int) error) error {
	delay := r.cfg.BaseDelay
	var lastErr error
	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn(attempt)
		if lastErr == nil {
			return nil
		}
		if attempt == r.cfg.MaxAttempts {
			break
		}
		if err := r.sleep(ctx, delay); err != nil {
			return err
		}
		delay = nextDelay(delay, r.cfg.Multiplier, r.cfg.MaxDelay)
	}
	return errors.Join(ErrMaxAttempts, lastErr)
}

func nextDelay(current time.Duration, multiplier float64, max time.Duration) time.Duration {
	next := time.Duration(float64(current) * multiplier)
	if next > max {
		return max
	}
	return next
}

func contextSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
