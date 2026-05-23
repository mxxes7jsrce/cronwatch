package jobregistry

import (
	"testing"
	"time"

	"github.com/cronwatch/internal/config"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func baseJobs() []config.Job {
	return []config.Job{
		{Name: "backup", Schedule: "@daily", MaxAge: 25 * time.Hour},
		{Name: "cleanup", Schedule: "@hourly", MaxAge: 90 * time.Minute},
	}
}

func TestNew_PopulatesEntries(t *testing.T) {
	reg := New(baseJobs())
	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestGet_KnownJob(t *testing.T) {
	reg := New(baseJobs())
	e, ok := reg.Get("backup")
	if !ok {
		t.Fatal("expected entry for 'backup'")
	}
	if e.Name != "backup" {
		t.Errorf("unexpected name: %s", e.Name)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	reg := New(baseJobs())
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Fatal("expected not found for unknown job")
	}
}

func TestTouch_UpdatesLastSeen(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(baseJobs(), fixedNow(now))

	ok := reg.Touch("backup")
	if !ok {
		t.Fatal("Touch should return true for known job")
	}
	e, _ := reg.Get("backup")
	if !e.LastSeen.Equal(now) {
		t.Errorf("expected LastSeen %v, got %v", now, e.LastSeen)
	}
}

func TestTouch_UnknownJobReturnsFalse(t *testing.T) {
	reg := New(baseJobs())
	if reg.Touch("ghost") {
		t.Fatal("Touch should return false for unknown job")
	}
}

func TestTouch_MultipleUpdatesLastSeen(t *testing.T) {
	first := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	second := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)

	reg := newWithClock(baseJobs(), fixedNow(first))
	reg.Touch("backup")

	reg.now = fixedNow(second)
	reg.Touch("backup")

	e, _ := reg.Get("backup")
	if !e.LastSeen.Equal(second) {
		t.Errorf("expected LastSeen to be updated to %v, got %v", second, e.LastSeen)
	}
}

func TestStale_ReturnsOverdueJobs(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(baseJobs(), fixedNow(base))

	// Touch backup 30 hours ago (exceeds 25h MaxAge)
	reg.entries["backup"].LastSeen = base.Add(-30 * time.Hour)
	// Touch cleanup 30 minutes ago (within 90m MaxAge)
	reg.entries["cleanup"].LastSeen = base.Add(-30 * time.Minute)

	stale := reg.Stale()
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(stale))
	}
	if stale[0].Name != "backup" {
		t.Errorf("expected 'backup' to be stale, got %s", stale[0].Name)
	}
}

func TestStale_ZeroLastSeenIgnored(t *testing.T) {
	reg := New(baseJobs()) // LastSeen is zero for all
	if len(reg.Stale()) != 0 {
		t.Fatal("expected no stale entries when LastSeen is zero")
	}
}
