package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stuffstash/stuff-stash/cmd/stuff-stash/internal/bootstrap"
	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/config"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	observer := observability.NewFanOut(observability.NewSlogObserver(logger))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		err = bootstrap.RunMigrationCommand(ctx, cfg, os.Args[2:], os.Stdout)
	} else {
		err = bootstrap.Run(ctx, cfg, observer)
	}
	if err != nil {
		bootstrap.RecordStartupFailure(observer, err)
		os.Exit(1)
	}
}
