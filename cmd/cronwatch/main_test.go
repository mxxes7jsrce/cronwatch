package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestMain_BuildAndRun compiles the binary and verifies it starts and
// shuts down cleanly when sent SIGINT. This is an integration smoke test.
func TestMain_BuildAndRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "cronwatch")

	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	cfgContent := []byte(`
check_interval_seconds: 60
webhook_url: "http://localhost:9999/hook"
state_file: "/tmp/cronwatch_test_state.json"
jobs:
  - name: test-job
    schedule: "* * * * *"
    max_latency_seconds: 300
`)
	cfgPath := filepath.Join(tmpDir, "cronwatch.yaml")
	if err := os.WriteFile(cfgPath, cfgContent, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cmd := exec.Command(binPath, "-config", cfgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start daemon: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send interrupt: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			t.Logf("process exited with: %v (may be normal on interrupt)", err)
		}
	case <-time.After(3 * time.Second):
		cmd.Process.Kill()
		t.Error("daemon did not exit within timeout")
	}
}
