package watcher_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/watcher"
)

func TestWatcher_IntegrationLargeFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "big.log")

	// Write more than 32KB of benign content followed by an error line.
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("create log: %v", err)
	}
	line := strings.Repeat("info: all systems nominal\n", 1400) // ~39KB
	f.WriteString(line)
	f.WriteString("FATAL: disk full\n")
	f.Close()

	job := config.Job{
		Name:            "big-job",
		LogFile:         logPath,
		FailureKeywords: []string{"FATAL"},
	}
	w, err := watcher.New(job, "big-job")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if res.Success {
		t.Error("expected failure for large file with FATAL in tail")
	}
	if res.MatchedKeyword != "FATAL" {
		t.Errorf("expected FATAL keyword, got %q", res.MatchedKeyword)
	}
}

func TestWatcher_IntegrationContextCancelled(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "job.log")
	os.WriteFile(logPath, []byte("ok\n"), 0o644)

	job := config.Job{Name: "ctx-job", LogFile: logPath}
	w, err := watcher.New(job, "ctx-job")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = w.Check(ctx)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}
