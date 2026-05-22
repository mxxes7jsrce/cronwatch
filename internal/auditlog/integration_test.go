package auditlog_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAuditLog_ConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer l.Close()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = l.Record(EventAlert, "job", "concurrent")
		}()
	}
	wg.Wait()

	f, _ := os.Open(path)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Errorf("invalid JSON on line %d: %v", count+1, err)
		}
		count++
	}
	if count != goroutines {
		t.Errorf("lines = %d, want %d", count, goroutines)
	}
}

func TestAuditLog_PersistsAcrossReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	l1, _ := New(path)
	_ = l1.Record(EventMissed, "nightly", "first open")
	l1.Close()

	l2, _ := New(path)
	_ = l2.Record(EventFailure, "nightly", "second open")
	l2.Close()

	f, _ := os.Open(path)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var events []Event
	for scanner.Scan() {
		var e Event
		_ = json.Unmarshal(scanner.Bytes(), &e)
		events = append(events, e)
	}
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	if events[0].Kind != EventMissed || events[1].Kind != EventFailure {
		t.Errorf("unexpected event kinds: %v, %v", events[0].Kind, events[1].Kind)
	}
}
