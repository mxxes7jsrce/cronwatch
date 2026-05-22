package auditlog

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempLogPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.log")
}

func TestRecord_WritesJSON(t *testing.T) {
	path := tempLogPath(t)
	l, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer l.Close()

	if err := l.Record(EventAlert, "backup", "missed run"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	f, _ := os.Open(path)
	defer f.Close()
	var e Event
	if err := json.NewDecoder(f).Decode(&e); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if e.Kind != EventAlert {
		t.Errorf("kind = %q, want %q", e.Kind, EventAlert)
	}
	if e.JobName != "backup" {
		t.Errorf("job = %q, want backup", e.JobName)
	}
	if e.Message != "missed run" {
		t.Errorf("message = %q, want 'missed run'", e.Message)
	}
}

func TestRecord_TimestampIsUTC(t *testing.T) {
	fixed := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	path := tempLogPath(t)
	l, _ := newWithClock(path, func() time.Time { return fixed })
	defer l.Close()

	_ = l.Record(EventSuccess, "sync", "")

	f, _ := os.Open(path)
	defer f.Close()
	var e Event
	_ = json.NewDecoder(f).Decode(&e)
	if !e.Timestamp.Equal(fixed) {
		t.Errorf("timestamp = %v, want %v", e.Timestamp, fixed)
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	path := tempLogPath(t)
	l, _ := New(path)
	defer l.Close()

	kinds := []EventKind{EventMissed, EventFailure, EventSuccess}
	for _, k := range kinds {
		if err := l.Record(k, "job", ""); err != nil {
			t.Fatalf("Record(%s): %v", k, err)
		}
	}

	f, _ := os.Open(path)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var got []EventKind
	for scanner.Scan() {
		var e Event
		_ = json.Unmarshal(scanner.Bytes(), &e)
		got = append(got, e.Kind)
	}
	if len(got) != 3 {
		t.Fatalf("entries = %d, want 3", len(got))
	}
	for i, k := range kinds {
		if got[i] != k {
			t.Errorf("entry[%d] = %q, want %q", i, got[i], k)
		}
	}
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/dir/audit.log")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}
