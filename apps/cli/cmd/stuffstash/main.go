package main

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/bootstrap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(bootstrap.Run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}
