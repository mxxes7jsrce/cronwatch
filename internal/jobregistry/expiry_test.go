package jobregistry

import (
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

func baseExpiryJobs() []config.Job {
	return []config.Job{
		{Name: "alpha", Schedule: "@every 1m", LogPath: "/tmp/alpha.log", Timeout: "2m"},
		{Name: "beta", Schedule: "@every 5m", LogPath: "/tmp/beta.log", Timeout: "10m"},
	}
}

func TestExpired_NoneExpired(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(baseExpiryJobs(), func() time.Time { return now })
	checker := NewExpiryChecker(reg)

	// Touch both jobs at "now"; check just 30 s later — neither should expire.
	reg.Touch("alpha")
	reg.Touch("beta")

	expired := checker.Expired(now.Add(30 * time.Second))
	if len(expired) != 0 {
		t.Fatalf("expected no expired jobs, got %v", expired)
	}
}

func TestExpired_OneExpired(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(baseExpiryJobs(), func() time.Time { return now })
	checker := NewExpiryChecker(reg)

	reg.Touch("alpha") // LastSeen = now, timeout 2 m → expires at now+2m
	reg.Touch("beta")  // LastSeen = now, timeout 10 m → expires at now+10m

	// Advance 3 minutes: alpha should be expired, beta should not.
	check := now.Add(3 * time.Minute)
	expired := checker.Expired(check)
	if len(expired) != 1 || expired[0] != "alpha" {
		t.Fatalf("expected [alpha], got %v", expired)
	}
}

func TestExpired_AllExpired(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(baseExpiryJobs(), func() time.Time { return now })
	checker := NewExpiryChecker(reg)

	reg.Touch("alpha")
	reg.Touch("beta")

	// Advance 15 minutes: both should be expired.
	expired := checker.Expired(now.Add(15 * time.Minute))
	if len(expired) != 2 {
		t.Fatalf("expected 2 expired jobs, got %v", expired)
	}
}

func TestNextExpiry_ReturnsEarliest(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	reg := newWithClock(baseExpiryJobs(), func() time.Time { return now })
	checker := NewExpiryChecker(reg)

	reg.Touch("alpha") // expires now+2m
	reg.Touch("beta")  // expires now+10m

	next, ok := checker.NextExpiry()
	if !ok {
		t.Fatal("expected ok=true")
	}
	want := now.Add(2 * time.Minute)
	if !next.Equal(want) {
		t.Fatalf("expected next expiry %v, got %v", want, next)
	}
}

func TestNextExpiry_EmptyRegistry(t *testing.T) {
	reg := newWithClock([]config.Job{}, time.Now)
	checker := NewExpiryChecker(reg)
	_, ok := checker.NextExpiry()
	if ok {
		t.Fatal("expected ok=false for empty registry")
	}
}
