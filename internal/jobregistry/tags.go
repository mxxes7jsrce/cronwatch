package jobregistry

import (
	"encoding/json"
	"net/http"
	"strings"
)

// TagsHandler returns an HTTP handler that exposes job listing filtered by tag.
// GET /jobs/tags?tag=<name> returns all jobs carrying that tag.
// GET /jobs/tags lists all distinct tags known to the registry.
func (r *Registry) TagsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tag := strings.TrimSpace(req.URL.Query().Get("tag"))

		if tag == "" {
			// Return all distinct tags.
			tags := r.allTags()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string][]string{"tags": tags})
			return
		}

		// Return jobs that carry the requested tag.
		names := r.jobsWithTag(tag)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string][]string{"jobs": names})
	}
}

// allTags returns a sorted, deduplicated slice of every tag present in the registry.
func (r *Registry) allTags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]struct{})
	for _, entry := range r.entries {
		for _, t := range entry.Job.Tags {
			seen[t] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sortStrings(out)
	return out
}

// jobsWithTag returns the names of all jobs that carry the given tag.
func (r *Registry) jobsWithTag(tag string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, entry := range r.entries {
		for _, t := range entry.Job.Tags {
			if t == tag {
				names = append(names, name)
				break
			}
		}
	}
	sortStrings(names)
	return names
}

// sortStrings sorts a string slice in-place (avoids importing sort everywhere).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
