package alertmanager

import (
	"sync"
	"time"
)

// rateLimiter enforces a maximum number of alerts per time window per job.
type rateLimiter struct {
	mu      sync.Mutex
	counts  map[string][]time.Time
	maxRate int
	window  time.Duration
	clock   func() time.Time
}

func newRateLimiter(maxRate int, window time.Duration, clock func() time.Time) *rateLimiter {
	if clock == nil {
		clock = time.Now
	}
	return &rateLimiter{
		counts:  make(map[string][]time.Time),
		maxRate: maxRate,
		window:  window,
		clock:   clock,
	}
}

// Allow returns true if the alert for the given key is within the rate limit.
func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	cutoff := now.Add(-r.window)

	times := r.counts[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.maxRate {
		r.counts[key] = filtered
		return false
	}

	r.counts[key] = append(filtered, now)
	return true
}

// Remaining returns how many more alerts are allowed for the given key
// within the current window.
func (r *rateLimiter) Remaining(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	cutoff := now.Add(-r.window)

	count := 0
	for _, t := range r.counts[key] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := r.maxRate - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
