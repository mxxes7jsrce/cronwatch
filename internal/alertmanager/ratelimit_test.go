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
		if !rl.Allow("job-a") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}

	if rl.Allow("job-a") {
		t.Fatal("expected Allow to return false after exceeding max rate")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	base := time.Now()
	clock := base
	rl := newRateLimiter(2, time.Minute, func() time.Time { return clock })

	rl.Allow("job-b")
	rl.Allow("job-b")

	if rl.Allow("job-b") {
		t.Fatal("expected rate limit to block third alert")
	}

	// Advance clock past window
	clock = base.Add(2 * time.Minute)

	if !rl.Allow("job-b") {
		t.Fatal("expected Allow after window reset")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	now := time.Now()
	rl := newRateLimiter(5, time.Minute, makeFixedClock(now))

	if got := rl.Remaining("job-c"); got != 5 {
		t.Fatalf("expected 5 remaining, got %d", got)
	}

	rl.Allow("job-c")
	rl.Allow("job-c")

	if got := rl.Remaining("job-c"); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
}

func TestRateLimiter_NilClockDefaultsToTimeNow(t *testing.T) {
	rl := newRateLimiter(1, time.Minute, nil)
	if !rl.Allow("job-d") {
		t.Fatal("expected first Allow with nil clock to succeed")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	now := time.Now()
	rl := newRateLimiter(1, time.Minute, makeFixedClock(now))

	if !rl.Allow("job-x") {
		t.Fatal("expected job-x to be allowed")
	}
	if !rl.Allow("job-y") {
		t.Fatal("expected job-y to be allowed independently")
	}
	if rl.Allow("job-x") {
		t.Fatal("expected job-x to be rate limited")
	}
}
