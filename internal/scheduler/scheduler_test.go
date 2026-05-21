package scheduler_test

import (
	"context"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/scheduler"
)

type mockRunner struct {
	calls atomic.Int32
}

func (m *mockRunner) Run(_ context.Context) error {
	m.calls.Add(1)
	return nil
}

func testCfg(intervalSec int) *config.Config {
	return &config.Config{
		CheckIntervalSeconds: intervalSec,
	}
}

func TestScheduler_RunsImmediately(t *testing.T) {
	runner := &mockRunner{}
	logger := log.New(os.Stderr, "", 0)
	s := scheduler.New(testCfg(60), runner, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	s.Start(ctx)

	if runner.calls.Load() < 1 {
		t.Error("expected at least one immediate run")
	}
}

func TestScheduler_TicksMultipleTimes(t *testing.T) {
	runner := &mockRunner{}
	logger := log.New(os.Stderr, "", 0)
	// Use a very short interval so we get multiple ticks in the test window.
	s := scheduler.New(testCfg(1), runner, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	s.Start(ctx)

	if runner.calls.Load() < 2 {
		t.Errorf("expected >=2 runs, got %d", runner.calls.Load())
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	runner := &mockRunner{}
	logger := log.New(os.Stderr, "", 0)
	s := scheduler.New(testCfg(60), runner, logger)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Error("scheduler did not stop after context cancellation")
	}
}
