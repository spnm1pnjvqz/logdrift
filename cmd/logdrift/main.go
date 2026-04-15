package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/logdrift/internal/config"
	"github.com/user/logdrift/internal/differ"
	"github.com/user/logdrift/internal/display"
	"github.com/user/logdrift/internal/runner"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "logdrift: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := "logdrift.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	r := runner.New(cfg.Services)
	channels, err := r.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting runners: %w", err)
	}
	defer r.StopAll()

	merged := runner.FanIn(ctx, channels...)

	pipeline := differ.NewPipeline(ctx, merged, cfg.DiffMode)

	printer := display.New(os.Stdout, cfg.DiffMode)
	printer.Run(ctx, pipeline)

	return nil
}
