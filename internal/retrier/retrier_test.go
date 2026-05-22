package retrier

import (
	"context"
	"errors"
	"testing"
	"time"
)

func instantSleep(_ context.Context, _ time.Duration) error { return nil }

func newFast(cfg Config) *Retrier {
	r := New(cfg)
	r.sleep = instantSleep
	return r
}

func TestRetrier_SucceedsOnFirstAttempt(t *testing.T) {
	r := newFast(DefaultConfig())
	calls := 0
	err := r.Do(context.Background(), func(_ int) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetrier_RetriesUpToMax(t *testing.T) {
	cfg := Config{MaxAttempts: 4, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond, Multiplier: 1}
	r := newFast(cfg)
	calls := 0
	sentinel := errors.New("boom")
	err := r.Do(context.Background(), func(_ int) error {
		calls++
		return sentinel
	})
	if !errors.Is(err, ErrMaxAttempts) {
		t.Fatalf("expected ErrMaxAttempts, got %v", err)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel wrapped, got %v", err)
	}
	if calls != 4 {
		t.Fatalf("expected 4 calls, got %d", calls)
	}
}

func TestRetrier_SucceedsOnRetry(t *testing.T) {
	r := newFast(DefaultConfig())
	calls := 0
	err := r.Do(context.Background(), func(_ int) error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetrier_StopsOnContextCancel(t *testing.T) {
	cfg := Config{MaxAttempts: 10, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second, Multiplier: 2}
	r := New(cfg) // real sleep so cancel can interrupt
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := r.Do(ctx, func(_ int) error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errors.New("fail")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestNextDelay_CapsAtMax(t *testing.T) {
	result := nextDelay(8*time.Second, 2.0, 10*time.Second)
	if result != 10*time.Second {
		t.Fatalf("expected 10s, got %v", result)
	}
}

func TestNextDelay_Multiplies(t *testing.T) {
	result := nextDelay(2*time.Second, 2.0, time.Minute)
	if result != 4*time.Second {
		t.Fatalf("expected 4s, got %v", result)
	}
}
