// Package watcher provides functionality for parsing and tailing system cron
// logs (or a configurable log file) to detect job execution events.
//
// It scans log lines for patterns matching configured job identifiers and
// reports successes or failures back to the state store via callbacks.
package watcher

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/user/cronwatch/internal/config"
)

// Event represents a detected cron job execution event parsed from a log line.
type Event struct {
	// JobName is the name of the job as defined in the config.
	JobName string
	// Success indicates whether the job completed without error.
	Success bool
	// Timestamp is when the event was detected.
	Timestamp time.Time
	// RawLine is the original log line that triggered the event.
	RawLine string
}

// Handler is a callback invoked when a job execution event is detected.
type Handler func(evt Event)

// Watcher tails a log file and emits Events for configured jobs.
type Watcher struct {
	cfg     *config.Config
	logPath string
	handler Handler
}

// New creates a new Watcher that reads from logPath and calls handler for
// each detected job event. logPath is typically /var/log/syslog or
// /var/log/cron depending on the host OS.
func New(cfg *config.Config, logPath string, handler Handler) (*Watcher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("watcher: config must not be nil")
	}
	if logPath == "" {
		return nil, fmt.Errorf("watcher: logPath must not be empty")
	}
	if handler == nil {
		return nil, fmt.Errorf("watcher: handler must not be nil")
	}
	return &Watcher{
		cfg:     cfg,
		logPath: logPath,
		handler: handler,
	}, nil
}

// Run opens the log file, seeks to the end, and continuously reads new lines
// until ctx is cancelled. It calls the handler for each matched event.
//
// Run blocks until the context is done.
func (w *Watcher) Run(ctx context.Context) error {
	f, err := os.Open(w.logPath)
	if err != nil {
		return fmt.Errorf("watcher: open log file %q: %w", w.logPath, err)
	}
	defer f.Close()

	// Seek to end so we only process new entries written after startup.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("watcher: seek log file: %w", err)
	}

	reader := bufio.NewReader(f)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("watcher: read log file: %w", err)
			}
			// No new data yet — wait briefly before retrying.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}

		w.processLine(strings.TrimRight(line, "\n"))
	}
}

// processLine checks the log line against each configured job's match pattern
// and fires the handler if a match is found.
func (w *Watcher) processLine(line string) {
	for _, job := range w.cfg.Jobs {
		if job.MatchPattern == "" {
			continue
		}
		if !strings.Contains(line, job.MatchPattern) {
			continue
		}

		// Determine success/failure by the absence of common error keywords.
		success := !containsFailureKeyword(line)

		w.handler(Event{
			JobName:   job.Name,
			Success:   success,
			Timestamp: time.Now().UTC(),
			RawLine:   line,
		})
	}
}

// containsFailureKeyword returns true if the log line contains any keyword
// that typically indicates a job failure.
func containsFailureKeyword(line string) bool {
	keywords := []string{"error", "fail", "exit code", "killed", "terminated"}
	lower := strings.ToLower(line)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
