package notify

import (
	"context"
	"fmt"
	"log"

	"github.com/cronwatch/cronwatch/internal/retrier"
)

// RetryingNotifier wraps a Notifier and retries failed sends according
// to the provided retrier.Config.
type RetryingNotifier struct {
	inner   Notifier
	retrier *retrier.Retrier
}

// Notifier is the interface satisfied by WebhookNotifier and any other
// alert delivery backend.
type Notifier interface {
	Send(ctx context.Context, payload map[string]any) error
}

// NewRetryingNotifier wraps inner with automatic retry behaviour.
func NewRetryingNotifier(inner Notifier, cfg retrier.Config) *RetryingNotifier {
	return &RetryingNotifier{
		inner:   inner,
		retrier: retrier.New(cfg),
	}
}

// Send attempts to deliver the payload, retrying on transient failures.
func (r *RetryingNotifier) Send(ctx context.Context, payload map[string]any) error {
	var lastAttempt int
	err := r.retrier.Do(ctx, func(attempt int) error {
		lastAttempt = attempt
		return r.inner.Send(ctx, payload)
	})
	if err != nil {
		log.Printf("retrying_notifier: delivery failed after %d attempt(s): %v", lastAttempt, err)
		return fmt.Errorf("notify: %w", err)
	}
	return nil
}
