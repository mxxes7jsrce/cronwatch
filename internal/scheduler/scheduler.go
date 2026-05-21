// Package scheduler is responsible for periodically triggering the monitor
// to check all configured cron jobs against their expected schedules.
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/user/cronwatch/internal/config"
)

// Runner is the interface satisfied by monitor.Monitor.
type Runner interface {
	Run(ctx context.Context) error
}

// Scheduler drives periodic monitor checks.
type Scheduler struct {
	cfg     *config.Config
	runner  Runner
	ticker  *time.Ticker
	logger  *log.Logger
}

// New creates a new Scheduler.
func New(cfg *config.Config, runner Runner, logger *log.Logger) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		runner: runner,
		logger: logger,
	}
}

// Start begins the scheduling loop, blocking until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	interval := time.Duration(s.cfg.CheckIntervalSeconds) * time.Second
	s.ticker = time.NewTicker(interval)
	defer s.ticker.Stop()

	s.logger.Printf("scheduler: starting with interval %s", interval)

	// Run immediately on start.
	s.runOnce(ctx)

	for {
		select {
		case <-s.ticker.C:
			s.runOnce(ctx)
		case <-ctx.Done():
			s.logger.Println("scheduler: shutting down")
			return
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) {
	if err := s.runner.Run(ctx); err != nil {
		s.logger.Printf("scheduler: monitor run error: %v", err)
	}
}
