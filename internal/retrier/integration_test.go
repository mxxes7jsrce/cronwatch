package retrier_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/retrier"
)

// TestRetrier_IntegrationHTTPFlaky simulates a flaky HTTP endpoint that
// fails twice before succeeding and verifies the retrier recovers.
func TestRetrier_IntegrationHTTPFlaky(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := retrier.Config{
		MaxAttempts: 5,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
	}
	r := retrier.New(cfg)

	err := r.Do(context.Background(), func(_ int) error {
		resp, err := http.Get(ts.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.New("non-200 response")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 HTTP calls, got %d", calls.Load())
	}
}

// TestRetrier_IntegrationExhausted verifies ErrMaxAttempts is returned
// when the endpoint never recovers.
func TestRetrier_IntegrationExhausted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := retrier.Config{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
		Multiplier:  2.0,
	}
	r := retrier.New(cfg)

	err := r.Do(context.Background(), func(_ int) error {
		resp, _ := http.Get(ts.URL)
		defer resp.Body.Close()
		return errors.New("server error")
	})
	if !errors.Is(err, retrier.ErrMaxAttempts) {
		t.Fatalf("expected ErrMaxAttempts, got %v", err)
	}
}
