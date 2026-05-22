package alertmanager_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cronwatch/internal/alertmanager"
	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/watcher"
)

type mockNotifier struct {
	mu     sync.Mutex
	calls  []string
	err    error
}

func (m *mockNotifier) Send(_ context.Context, msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, msg)
	return m.err
}

func (m *mockNotifier) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func baseJobCfg(name string) config.Job {
	return config.Job{
		Name:          name,
		Schedule:      "@hourly",
		LogFile:       "/var/log/" + name + ".log",
		Cooldown:      60,
		FailKeywords:  []string{"error"},
	}
}

func TestAlertManager_SendsOnFailure(t *testing.T) {
	n := &mockNotifier{}
	am := alertmanager.New(n, 0)

	res := watcher.Result{Job: baseJobCfg("backup"), Failed: true, Reason: "keyword: error"}
	am.Handle(context.Background(), res)

	if n.CallCount() != 1 {
		t.Fatalf("expected 1 alert, got %d", n.CallCount())
	}
}

func TestAlertManager_DeduplicatesWithinCooldown(t *testing.T) {
	n := &mockNotifier{}
	am := alertmanager.New(n, 5*time.Minute)

	res := watcher.Result{Job: baseJobCfg("backup"), Failed: true, Reason: "keyword: error"}
	am.Handle(context.Background(), res)
	am.Handle(context.Background(), res)

	if n.CallCount() != 1 {
		t.Fatalf("expected 1 alert due to dedup, got %d", n.CallCount())
	}
}

func TestAlertManager_NoAlertOnSuccess(t *testing.T) {
	n := &mockNotifier{}
	am := alertmanager.New(n, 0)

	res := watcher.Result{Job: baseJobCfg("backup"), Failed: false}
	am.Handle(context.Background(), res)

	if n.CallCount() != 0 {
		t.Fatalf("expected 0 alerts on success, got %d", n.CallCount())
	}
}
