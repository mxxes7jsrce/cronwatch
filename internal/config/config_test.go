package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
check_interval: 30s
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout: 10m
    alert_on: [missed, failed]
alerts:
  slack:
    webhook_url: https://hooks.slack.com/test
`
	path := writeTempConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected check_interval 30s, got %v", cfg.CheckInterval)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
}

func TestLoad_DefaultCheckInterval(t *testing.T) {
	yaml := `
jobs:
  - name: cleanup
    schedule: "*/5 * * * *"
`
	path := writeTempConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != time.Minute {
		t.Errorf("expected default check_interval 1m, got %v", cfg.CheckInterval)
	}
}

func TestLoad_NoJobs(t *testing.T) {
	yaml := `check_interval: 1m\njobs: []\n`
	path := writeTempConfig(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty jobs, got nil")
	}
}

func TestLoad_DuplicateJobName(t *testing.T) {
	yaml := `
jobs:
  - name: sync
    schedule: "0 * * * *"
  - name: sync
    schedule: "0 * * * *"
`
	path := writeTempConfig(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate job name, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
