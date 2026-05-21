package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/cronwatch/internal/config"
)

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "test.log")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempLog: %v", err)
	}
	return p
}

func baseJob(logPath string) config.Job {
	return config.Job{
		Name:            "test-job",
		LogFile:         logPath,
		FailureKeywords: []string{"ERROR", "FATAL"},
	}
}

func TestCheck_SuccessNoKeywords(t *testing.T) {
	logPath := writeTempLog(t, "job completed successfully\n")
	w, err := New(baseJob(logPath), "test-job")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success, got failure: %s", res.Output)
	}
}

func TestCheck_FailureKeywordDetected(t *testing.T) {
	logPath := writeTempLog(t, "starting job\nERROR: something went wrong\n")
	w, err := New(baseJob(logPath), "test-job")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if res.Success {
		t.Error("expected failure due to keyword, got success")
	}
	if res.MatchedKeyword != "ERROR" {
		t.Errorf("expected matched keyword ERROR, got %q", res.MatchedKeyword)
	}
}

func TestCheck_MissingLogFile(t *testing.T) {
	job := config.Job{
		Name:    "missing-job",
		LogFile: "/nonexistent/path/job.log",
	}
	w, err := New(job, "missing-job")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = w.Check(context.Background())
	if err == nil {
		t.Error("expected error for missing log file, got nil")
	}
}

func TestContainsFailureKeyword(t *testing.T) {
	cases := []struct {
		output   string
		keywords []string
		want     string
	}{
		{"all good", []string{"ERROR"}, ""},
		{"FATAL crash", []string{"ERROR", "FATAL"}, "FATAL"},
		{"error in lowercase", []string{"error"}, "error"},
	}
	for _, tc := range cases {
		got := containsFailureKeyword(tc.output, tc.keywords)
		if got != tc.want {
			t.Errorf("containsFailureKeyword(%q, %v) = %q; want %q", tc.output, tc.keywords, got, tc.want)
		}
	}
}
