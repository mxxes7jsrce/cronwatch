package alertmanager

import (
	"context"
	"testing"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/watcher"
)

// mockNotifier records calls to Send.
type mockNotifier struct {
	calls []string
	err   error
}

func (m *mockNotifier) Send(_ context.Context, msg string) error {
	m.calls = append(m.calls, msg)
	return m.err
}

var baseJobCfg = config.Job{
	Name:    "backup",
	LogPath: "/var/log/backup.log",
}

func baseCfg(cooldown int) config.Config {
	return config.Config{AlertCooldownMinutes: cooldown}
}

func TestAlertManager_SendsOnFailure(t *testing.T) {
	n := &mockNotifier{}
	am := newWithClock(baseCfg(5), n, time.Now)

	result := watcher.Result{Healthy: false, MatchedKeyword: "error"}
	am.Evaluate(context.Background(), baseJobCfg, result)

	if len(n.calls) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(n.calls))
	}
	if want := "[cronwatch] FAILURE"; len(n.calls[0]) == 0 {
		t.Errorf("expected message containing %q, got empty", want)
	}
}

func TestAlertManager_DeduplicatesWithinCooldown(t *testing.T) {
	n := &mockNotifier{}
	now := time.Now()
	clock := func() time.Time { return now }
	am := newWithClock(baseCfg(10), n, clock)

	result := watcher.Result{Healthy: false, MatchedKeyword: "error"}
	am.Evaluate(context.Background(), baseJobCfg, result)
	am.Evaluate(context.Background(), baseJobCfg, result)

	if len(n.calls) != 1 {
		t.Errorf("expected 1 alert due to dedup, got %d", len(n.calls))
	}
}

func TestAlertManager_AllowsAlertAfterCooldown(t *testing.T) {
	n := &mockNotifier{}
	now := time.Now()
	clock := func() time.Time { return now }
	am := newWithClock(baseCfg(5), n, clock)

	result := watcher.Result{Healthy: false, MatchedKeyword: "error"}
	am.Evaluate(context.Background(), baseJobCfg, result)

	// Advance clock past cooldown.
	now = now.Add(6 * time.Minute)
	am.Evaluate(context.Background(), baseJobCfg, result)

	if len(n.calls) != 2 {
		t.Errorf("expected 2 alerts after cooldown, got %d", len(n.calls))
	}
}

func TestAlertManager_SkipsHealthyResult(t *testing.T) {
	n := &mockNotifier{}
	am := newWithClock(baseCfg(5), n, time.Now)

	result := watcher.Result{Healthy: true}
	am.Evaluate(context.Background(), baseJobCfg, result)

	if len(n.calls) != 0 {
		t.Errorf("expected no alerts for healthy result, got %d", len(n.calls))
	}
}
