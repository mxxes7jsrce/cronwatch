package alertmanager

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestDedupStore_AllowsFirstAlert(t *testing.T) {
	now := time.Now()
	d := newDedupStore(fixedClock(now))

	if !d.allow("backup", "missed", 5*time.Minute) {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestDedupStore_SuppressesWithinCooldown(t *testing.T) {
	now := time.Now()
	d := newDedupStore(fixedClock(now))

	d.allow("backup", "missed", 5*time.Minute)

	if d.allow("backup", "missed", 5*time.Minute) {
		t.Fatal("expected alert to be suppressed within cooldown")
	}
}

func TestDedupStore_AllowsAfterCooldown(t *testing.T) {
	now := time.Now()
	d := newDedupStore(fixedClock(now))

	d.allow("backup", "missed", 5*time.Minute)

	// Advance clock beyond cooldown.
	d.clock = fixedClock(now.Add(6 * time.Minute))

	if !d.allow("backup", "missed", 5*time.Minute) {
		t.Fatal("expected alert to be allowed after cooldown")
	}
}

func TestDedupStore_DifferentKindsAreIndependent(t *testing.T) {
	now := time.Now()
	d := newDedupStore(fixedClock(now))

	d.allow("backup", "missed", 5*time.Minute)

	if !d.allow("backup", "failed", 5*time.Minute) {
		t.Fatal("expected different kind to be allowed independently")
	}
}

func TestDedupStore_Reset(t *testing.T) {
	now := time.Now()
	d := newDedupStore(fixedClock(now))

	d.allow("backup", "missed", 5*time.Minute)
	d.reset("backup", "missed")

	if !d.allow("backup", "missed", 5*time.Minute) {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestDedupStore_DifferentJobsAreIndependent(t *testing.T) {
	now := time.Now()
	d := newDedupStore(fixedClock(now))

	d.allow("backup", "missed", 5*time.Minute)

	if !d.allow("sync", "missed", 5*time.Minute) {
		t.Fatal("expected different job to be allowed independently")
	}
}
