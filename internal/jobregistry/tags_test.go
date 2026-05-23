package jobregistry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

var tagJobs = []config.Job{
	{Name: "backup", Schedule: "@daily", Tags: []string{"infra", "storage"}},
	{Name: "report", Schedule: "@weekly", Tags: []string{"infra", "analytics"}},
	{Name: "cleanup", Schedule: "@hourly", Tags: []string{"storage"}},
	{Name: "ping", Schedule: "@every 5m"},
}

func newTagRegistry(t *testing.T) *Registry {
	t.Helper()
	fixed := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	r, err := newWithClock(tagJobs, func() time.Time { return fixed })
	if err != nil {
		t.Fatalf("newWithClock: %v", err)
	}
	return r
}

func TestTags_ListAllTags(t *testing.T) {
	r := newTagRegistry(t)
	req := httptest.NewRequest(http.MethodGet, "/jobs/tags", nil)
	rec := httptest.NewRecorder()
	r.TagsHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	want := []string{"analytics", "infra", "storage"}
	if !reflect.DeepEqual(body["tags"], want) {
		t.Errorf("tags = %v; want %v", body["tags"], want)
	}
}

func TestTags_FilterByTag(t *testing.T) {
	r := newTagRegistry(t)
	req := httptest.NewRequest(http.MethodGet, "/jobs/tags?tag=infra", nil)
	rec := httptest.NewRecorder()
	r.TagsHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	want := []string{"backup", "report"}
	if !reflect.DeepEqual(body["jobs"], want) {
		t.Errorf("jobs = %v; want %v", body["jobs"], want)
	}
}

func TestTags_FilterByTag_NoMatch(t *testing.T) {
	r := newTagRegistry(t)
	req := httptest.NewRequest(http.MethodGet, "/jobs/tags?tag=nonexistent", nil)
	rec := httptest.NewRecorder()
	r.TagsHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body["jobs"]) != 0 {
		t.Errorf("expected empty jobs list, got %v", body["jobs"])
	}
}

func TestTags_MethodNotAllowed(t *testing.T) {
	r := newTagRegistry(t)
	req := httptest.NewRequest(http.MethodPost, "/jobs/tags", nil)
	rec := httptest.NewRecorder()
	r.TagsHandler()(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestTags_ContentTypeJSON(t *testing.T) {
	r := newTagRegistry(t)
	req := httptest.NewRequest(http.MethodGet, "/jobs/tags", nil)
	rec := httptest.NewRecorder()
	r.TagsHandler()(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
}
