package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/monitor"
	"github.com/user/cronwatch/internal/notify"
	"github.com/user/cronwatch/internal/scheduler"
	"github.com/user/cronwatch/internal/state"
)

func main() {
	cfgPath := flag.String("config", "configs/cronwatch.yaml", "path to config file")
	flag.Parse()

	logger := log.New(os.Stdout, "cronwatch: ", log.LstdFlags)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	store, err := state.NewStore(cfg.StateFile)
	if err != nil {
		logger.Fatalf("failed to open state store: %v", err)
	}

	notifier, err := notify.NewWebhookNotifier(cfg.WebhookURL, 0)
	if err != nil {
		logger.Fatalf("failed to create notifier: %v", err)
	}

	mon := monitor.New(cfg, store, notifier, logger)
	sch := scheduler.New(cfg, mon, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Println("starting cronwatch daemon")
	sch.Start(ctx)
	logger.Println("cronwatch stopped")
}
