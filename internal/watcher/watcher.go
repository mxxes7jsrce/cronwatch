package watcher

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/user/cronwatch/internal/config"
)

const maxReadBytes = 32 * 1024 // 32 KB tail read limit

// Watcher checks a cron job's log file for signs of failure.
type Watcher struct {
	job     config.Job
	jobName string
}

// New creates a new Watcher for the given job configuration.
func New(job config.Job, jobName string) (*Watcher, error) {
	if jobName == "" {
		return nil, fmt.Errorf("watcher: jobName must not be empty")
	}
	return &Watcher{job: job, jobName: jobName}, nil
}

// Check reads the configured log file and returns a Result indicating
// whether the job appears to have succeeded or failed.
func (w *Watcher) Check(ctx context.Context) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	output, err := w.readLog()
	if err != nil {
		return Result{}, fmt.Errorf("watcher %q: %w", w.jobName, err)
	}

	kw := containsFailureKeyword(output, w.job.FailureKeywords)
	return Result{
		JobName:        w.jobName,
		Success:        kw == "",
		Output:         output,
		MatchedKeyword: kw,
		CheckedAt:      time.Now().UTC(),
	}, nil
}

// readLog reads up to maxReadBytes from the tail of the log file.
func (w *Watcher) readLog() (string, error) {
	f, err := os.Open(w.job.LogFile)
	if err != nil {
		return "", fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("stat log file: %w", err)
	}

	size := info.Size()
	offset := int64(0)
	if size > maxReadBytes {
		offset = size - maxReadBytes
	}

	if _, err := f.Seek(offset, 0); err != nil {
		return "", fmt.Errorf("seek log file: %w", err)
	}

	buf := make([]byte, size-offset)
	n, err := f.Read(buf)
	if err != nil {
		return "", fmt.Errorf("read log file: %w", err)
	}
	return string(buf[:n]), nil
}

// containsFailureKeyword scans output for any of the provided keywords.
// Returns the first matched keyword, or empty string if none found.
func containsFailureKeyword(output string, keywords []string) string {
	for _, kw := range keywords {
		if strings.Contains(output, kw) {
			return kw
		}
	}
	return ""
}
