// Package jobregistry — staleness sub-feature.
//
// The staleness report identifies jobs that have not sent a heartbeat within
// their configured max interval. This is distinct from the expiry checker,
// which fires a callback; the staleness handler simply exposes the current
// state over HTTP so operators and dashboards can query it on demand.
//
// Typical usage:
//
//	reg := jobregistry.New(cfg, clock)
//	http.Handle("/jobs/stale", reg.StalenessHandler())
//
// The JSON response includes each stale job's name, last-seen timestamp,
// how long it has been silent, and its configured max interval, making it
// straightforward to build alerting rules on top.
package jobregistry
