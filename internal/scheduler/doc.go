// Package scheduler provides a time-driven loop that invokes the monitor
// at a configurable interval defined in the application config.
//
// Usage:
//
//	// Build dependencies.
//	 mon := monitor.New(cfg, store, notifier, log.Default())
//	 sch := scheduler.New(cfg, mon, log.Default())
//
//	 // Block until ctx is cancelled (e.g. on SIGTERM).
//	 sch.Start(ctx)
//
// The scheduler always performs an initial check immediately on Start,
// then repeats at cfg.CheckIntervalSeconds intervals.
package scheduler
