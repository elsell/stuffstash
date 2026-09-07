package bootstrap

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func startProfiling(ctx context.Context, observer ports.Observer) (func(), error) {
	cfg, err := config.LoadProfiling()
	if err != nil {
		return nil, err
	}
	profiler, err := observability.NewProfiler(ctx, cfg, observer)
	if err != nil {
		return nil, err
	}
	return func() {
		shutdown, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout+time.Second)
		defer cancel()
		if err := profiler.Stop(shutdown); err != nil {
			observer.Record(context.Background(), ports.Event{Name: ports.EventApplicationShutdownFailed, Message: "profiling shutdown incomplete"})
		}
	}, nil
}
