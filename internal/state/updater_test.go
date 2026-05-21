package state

import (
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(tempPath(t))
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return s
}

func TestRecordSuccess(t *testing.T) {
	s := newTestStore(t)
	_ = RecordFailure(s, "job1") // pre-populate with failure
	_ = RecordSuccess(s, "job1")

	js := s.Get("job1")
	if js.LastStatus != "ok" {
		t.Errorf("expected ok, got %s", js.LastStatus)
	}
	if js.FailCount != 0 {
		t.Errorf("expected FailCount reset to 0, got %d", js.FailCount)
	}
}

func TestRecordFailure_Increments(t *testing.T) {
	s := newTestStore(t)
	_ = RecordFailure(s, "job2")
	_ = RecordFailure(s, "job2")

	js := s.Get("job2")
	if js.FailCount != 2 {
		t.Errorf("expected FailCount 2, got %d", js.FailCount)
	}
}

func TestRecordMissed_Increments(t *testing.T) {
	s := newTestStore(t)
	_ = RecordMissed(s, "job3")
	_ = RecordMissed(s, "job3")

	js := s.Get("job3")
	if js.MissedCount != 2 {
		t.Errorf("expected MissedCount 2, got %d", js.MissedCount)
	}
	if js.LastStatus != "missed" {
		t.Errorf("expected status missed, got %s", js.LastStatus)
	}
}

func TestIsMissed_True(t *testing.T) {
	js := &JobState{
		Name:     "job4",
		LastSeen: time.Now().UTC().Add(-2 * time.Hour),
	}
	if !IsMissed(js, time.Hour, 5*time.Minute) {
		t.Error("expected job to be missed")
	}
}

func TestIsMissed_False(t *testing.T) {
	js := &JobState{
		Name:     "job5",
		LastSeen: time.Now().UTC().Add(-30 * time.Minute),
	}
	if IsMissed(js, time.Hour, 5*time.Minute) {
		t.Error("expected job to not be missed")
	}
}

func TestIsMissed_NilState(t *testing.T) {
	if IsMissed(nil, time.Hour, time.Minute) {
		t.Error("nil state should not be considered missed")
	}
}
