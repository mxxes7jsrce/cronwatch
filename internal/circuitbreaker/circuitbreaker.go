// Package circuitbreaker implements a simple circuit breaker that prevents
// repeated alert delivery attempts when a notifier is consistently failing.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking calls
	StateHalfOpen              // testing if service recovered
)

// ErrCircuitOpen is returned when the circuit is open and calls are blocked.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Breaker is a circuit breaker that trips after a threshold of consecutive
// failures and resets after a configurable timeout.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	threshold    int
	resetTimeout time.Duration
	lastFailure  time.Time
	clock        func() time.Time
}

// New creates a new Breaker with the given failure threshold and reset timeout.
func New(threshold int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
		clock:        time.Now,
	}
}

// newWithClock creates a Breaker with an injectable clock for testing.
func newWithClock(threshold int, resetTimeout time.Duration, clock func() time.Time) *Breaker {
	b := New(threshold, resetTimeout)
	b.clock = clock
	return b
}

// Allow reports whether the call should be allowed through.
// It transitions an open circuit to half-open if the reset timeout has elapsed.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateHalfOpen:
		return true
	case StateOpen:
		if b.clock().Sub(b.lastFailure) >= b.resetTimeout {
			b.state = StateHalfOpen
			return true
		}
		return false
	}
	return false
}

// RecordSuccess records a successful call, resetting the breaker to closed.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure records a failed call. After threshold failures the circuit opens.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	b.lastFailure = b.clock()
	if b.failures >= b.threshold {
		b.state = StateOpen
	}
}

// State returns the current state of the circuit breaker.
func (b *Breaker) CurrentState() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
