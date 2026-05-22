package circuitbreaker

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestBreaker_ClosedByDefault(t *testing.T) {
	b := New(3, 10*time.Second)
	if !b.Allow() {
		t.Fatal("expected Allow() == true for a fresh breaker")
	}
	if b.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed, got %v", b.CurrentState())
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	now := time.Now()
	b := newWithClock(3, 10*time.Second, fixedClock(now))

	b.RecordFailure()
	b.RecordFailure()
	if b.CurrentState() != StateClosed {
		t.Fatal("should still be closed after 2 failures with threshold 3")
	}
	b.RecordFailure()
	if b.CurrentState() != StateOpen {
		t.Fatalf("expected StateOpen after threshold, got %v", b.CurrentState())
	}
}

func TestBreaker_BlocksWhenOpen(t *testing.T) {
	now := time.Now()
	b := newWithClock(1, 30*time.Second, fixedClock(now))
	b.RecordFailure()

	if b.Allow() {
		t.Fatal("expected Allow() == false when circuit is open")
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	base := time.Now()
	b := newWithClock(1, 5*time.Second, fixedClock(base))
	b.RecordFailure()

	// advance clock past reset timeout
	b.clock = fixedClock(base.Add(6 * time.Second))

	if !b.Allow() {
		t.Fatal("expected Allow() == true after reset timeout (half-open)")
	}
	if b.CurrentState() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", b.CurrentState())
	}
}

func TestBreaker_SuccessResetsClosed(t *testing.T) {
	base := time.Now()
	b := newWithClock(1, 5*time.Second, fixedClock(base))
	b.RecordFailure()

	b.clock = fixedClock(base.Add(6 * time.Second))
	b.Allow() // transitions to half-open
	b.RecordSuccess()

	if b.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed after success, got %v", b.CurrentState())
	}
	if !b.Allow() {
		t.Fatal("expected Allow() == true after reset")
	}
}

func TestBreaker_FailureInHalfOpenReopens(t *testing.T) {
	base := time.Now()
	b := newWithClock(1, 5*time.Second, fixedClock(base))
	b.RecordFailure()

	b.clock = fixedClock(base.Add(6 * time.Second))
	b.Allow() // half-open
	b.RecordFailure()

	if b.CurrentState() != StateOpen {
		t.Fatalf("expected StateOpen after failure in half-open, got %v", b.CurrentState())
	}
}
