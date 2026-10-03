package main

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/bootstrap"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	os.Exit(bootstrap.Run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}
