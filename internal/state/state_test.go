package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestNewStore_EmptyFile(t *testing.T) {
	s, err := NewStore(tempPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := s.All(); len(got) != 0 {
		t.Errorf("expected empty store, got %d entries", len(got))
	}
}

func TestSetAndGet(t *testing.T) {
	path := tempPath(t)
	s, _ := NewStore(path)

	js := &JobState{
		Name:       "backup",
		LastSeen:   time.Now().UTC().Truncate(time.Second),
		LastStatus: "ok",
	}
	if err := s.Set(js); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got := s.Get("backup")
	if got == nil {
		t.Fatal("expected state, got nil")
	}
	if got.LastStatus != "ok" {
		t.Errorf("expected status ok, got %s", got.LastStatus)
	}
}

func TestPersistence(t *testing.T) {
	path := tempPath(t)
	s1, _ := NewStore(path)
	_ = s1.Set(&JobState{Name: "cleanup", LastStatus: "failed", FailCount: 3})

	s2, err := NewStore(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	got := s2.Get("cleanup")
	if got == nil {
		t.Fatal("state not persisted")
	}
	if got.FailCount != 3 {
		t.Errorf("expected FailCount 3, got %d", got.FailCount)
	}
}

func TestGet_Missing(t *testing.T) {
	s, _ := NewStore(tempPath(t))
	if got := s.Get("nonexistent"); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestNewStore_CorruptFile(t *testing.T) {
	path := tempPath(t)
	_ = os.WriteFile(path, []byte("not json{"), 0644)
	_, err := NewStore(path)
	if err == nil {
		t.Error("expected error for corrupt state file")
	}
}
