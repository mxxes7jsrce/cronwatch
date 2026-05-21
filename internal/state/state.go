package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// JobState holds the last known execution state of a cron job.
type JobState struct {
	Name        string    `json:"name"`
	LastSeen    time.Time `json:"last_seen"`
	LastStatus  string    `json:"last_status"` // "ok", "failed", "missed"
	FailCount   int       `json:"fail_count"`
	MissedCount int       `json:"missed_count"`
}

// Store manages persisted job states.
type Store struct {
	mu     sync.RWMutex
	states map[string]*JobState
	path   string
}

// NewStore creates a Store backed by the given file path.
func NewStore(path string) (*Store, error) {
	s := &Store{
		states: make(map[string]*JobState),
		path:   path,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Get returns the state for a job by name, or nil if not found.
func (s *Store) Get(name string) *JobState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[name]
}

// Set updates the state for a job and persists to disk.
func (s *Store) Set(js *JobState) error {
	s.mu.Lock()
	s.states[js.Name] = js
	s.mu.Unlock()
	return s.save()
}

// All returns a snapshot of all job states.
func (s *Store) All() []*JobState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*JobState, 0, len(s.states))
	for _, v := range s.states {
		copy := *v
		out = append(out, &copy)
	}
	return out
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var states []*JobState
	if err := json.Unmarshal(data, &states); err != nil {
		return err
	}
	for _, js := range states {
		s.states[js.Name] = js
	}
	return nil
}

func (s *Store) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]*JobState, 0, len(s.states))
	for _, v := range s.states {
		states = append(states, v)
	}
	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
