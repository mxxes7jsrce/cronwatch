package alertmanager

import (
	"sync"
	"time"
)

// rateLimiter enforces a maximum number of alerts per rolling time window.
type rateLimiter struct {
	mu        sync.Mutex
	max       int
	window    time.Duration
	timestamps []time.Time
	clock     func() time.Time
}

func newRateLimiter(max int, window time.Duration, clock func() time.Time) *rateLimiter {
	if clock == nil {
		clock = time.Now
	}
	return &rateLimiter{
		max:    max,
		window: window,
		clock:  clock,
	}
}

// Allow returns true if the alert is within the rate limit budget.
func (r *rateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	cutoff := now.Add(-r.window)

	// Evict timestamps outside the window.
	active := r.timestamps[:0]
	for _, t := range r.timestamps {
		if t.After(cutoff) {
			active = append(active, t)
		}
	}
	r.timestamps = active

	if len(r.timestamps) >= r.max {
		return false
	}

	r.timestamps = append(r.timestamps, now)
	return true
}

// Remaining returns how many more alerts are allowed in the current window.
func (r *rateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	cutoff := now.Add(-r.window)

	count := 0
	for _, t := range r.timestamps {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := r.max - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
