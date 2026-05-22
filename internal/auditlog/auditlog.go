// Package auditlog provides a structured append-only audit trail of
// cronwatch alerting and monitoring events written to a rotating log file.
package auditlog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// EventKind classifies the type of audit event.
type EventKind string

const (
	EventAlert   EventKind = "alert"
	EventMissed  EventKind = "missed"
	EventFailure EventKind = "failure"
	EventSuccess EventKind = "success"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	JobName   string    `json:"job_name"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes audit events as newline-delimited JSON to a file.
type Logger struct {
	mu   sync.Mutex
	file *os.File
	now  func() time.Time
}

// New opens (or creates) the audit log file at path and returns a Logger.
func New(path string) (*Logger, error) {
	return newWithClock(path, time.Now)
}

func newWithClock(path string, now func() time.Time) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("auditlog: open %q: %w", path, err)
	}
	return &Logger{file: f, now: now}, nil
}

// Record writes an event to the audit log.
func (l *Logger) Record(kind EventKind, jobName, message string) error {
	e := Event{
		Timestamp: l.now().UTC(),
		Kind:      kind,
		JobName:   jobName,
		Message:   message,
	}
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("auditlog: marshal: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = fmt.Fprintf(l.file, "%s\n", data)
	return err
}

// Close flushes and closes the underlying file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
