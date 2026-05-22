package alertmanager

import (
	"testing"
	"time"
)

func makeFixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRateLimiter_AllowsUpToMax(t *testing.T) {
	now := time.Now()
	rl := newRateLimiter(3, time.Minute, makeFixedClock(now))

	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Fatalf("expected Allow() == true on call %d", i+1)
		}
	}

	if rl.Allow() {
		t.Fatal("expected Allow() == false after max reached")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	base := time.Now()
	current := base
	clock := func() time.Time { return current }

	rl := newRateLimiter(2, time.Minute, clock)
	rl.Allow()
	rl.Allow()

	if rl.Allow() {
		t.Fatal("expected rate limit to be hit")
	}

	// Advance time past the window.
	current = base.Add(61 * time.Second)

	if !rl.Allow() {
		t.Fatal("expected Allow() == true after window elapsed")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	now := time.Now()
	rl := newRateLimiter(5, time.Minute, makeFixedClock(now))

	if got := rl.Remaining(); got != 5 {
		t.Fatalf("expected 5 remaining, got %d", got)
	}

	rl.Allow()
	rl.Allow()

	if got := rl.Remaining(); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
}

func TestRateLimiter_NilClockDefaultsToTimeNow(t *testing.T) {
	rl := newRateLimiter(1, time.Minute, nil)
	if !rl.Allow() {
		t.Fatal("expected first Allow() to succeed with nil clock")
	}
}

func TestRateLimiter_RemainingNeverNegative(t *testing.T) {
	now := time.Now()
	rl := newRateLimiter(1, time.Minute, makeFixedClock(now))
	rl.Allow()
	rl.Allow() // exceeds max

	if got := rl.Remaining(); got != 0 {
		t.Fatalf("expected 0 remaining, got %d", got)
	}
}
