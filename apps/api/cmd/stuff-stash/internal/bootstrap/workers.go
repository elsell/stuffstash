package bootstrap

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func startOutboxWorkers(ctx context.Context, application app.App, observer ports.Observer, cfg config.Config) {
	go drainAuthorizationOutbox(ctx, application, observer, cfg.AuthorizationOutboxDrainLimit, cfg.AuthorizationOutboxDrainInterval)
	go drainBlobDeletionOutbox(ctx, application, observer, cfg.BlobDeletionOutboxDrainLimit, cfg.BlobDeletionOutboxDrainInterval)
}

func drainAuthorizationOutbox(ctx context.Context, application app.App, observer ports.Observer, limit int, interval time.Duration) {
	drain := func() {
		if err := application.DrainAuthorizationOutbox(ctx, limit); err != nil {
			observer.Record(ctx, ports.Event{
				Name:    ports.EventAuthorizationOutboxFailed,
				Message: "authorization outbox background drain failed",
				Fields:  map[string]string{"error": err.Error()},
			})
		}
	}
	runPeriodicDrain(ctx, interval, drain)
}

func drainBlobDeletionOutbox(ctx context.Context, application app.App, observer ports.Observer, limit int, interval time.Duration) {
	drain := func() {
		if err := application.DrainBlobDeletionOutbox(ctx, limit); err != nil {
			observer.Record(ctx, ports.Event{
				Name:    ports.EventBlobDeletionOutboxFailed,
				Message: "blob deletion outbox background drain failed",
				Fields:  map[string]string{"error": err.Error()},
			})
		}
	}
	runPeriodicDrain(ctx, interval, drain)
}

func runPeriodicDrain(ctx context.Context, interval time.Duration, drain func()) {
	drain()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			drain()
		}
	}
}
