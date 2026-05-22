package alertmanager

import (
	"sync"
	"time"
)

// rateLimiter enforces a maximum number of alerts per window across all jobs.
// It is used to prevent alert storms when many jobs fail simultaneously.
type rateLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	clock    func() time.Time
	buckets  []time.Time
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

// Allow returns true if an alert may be sent, false if the rate limit is exceeded.
// It prunes expired entries before checking.
func (r *rateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	cutoff := now.Add(-r.window)

	// Prune entries outside the window.
	valid := r.buckets[:0]
	for _, t := range r.buckets {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	r.buckets = valid

	if len(r.buckets) >= r.max {
		return false
	}

	r.buckets = append(r.buckets, now)
	return true
}

// Remaining returns the number of alerts still allowed in the current window.
func (r *rateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	cutoff := now.Add(-r.window)
	count := 0
	for _, t := range r.buckets {
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
