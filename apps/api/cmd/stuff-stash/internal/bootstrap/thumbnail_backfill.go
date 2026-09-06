package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func startThumbnailBackfill(ctx context.Context, backfill ports.ThumbnailBackfill, clock ports.Clock, observer ports.Observer, cfg config.ThumbnailConfig) func() {
	if !cfg.WorkerEnabled || !cfg.BackfillEnabled {
		return func() {}
	}
	lifetime, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for lifetime.Err() == nil {
			batch, end := context.WithTimeout(lifetime, cfg.LeaseDuration)
			progress, err := backfill.BackfillThumbnailJobs(batch, cfg.BackfillBatchSize, clock.Now())
			end()
			if lifetime.Err() != nil {
				return
			}
			if err != nil {
				observer.Record(lifetime, ports.Event{Name: ports.EventThumbnailWorkerFailed, Message: "thumbnail backfill failed", Fields: map[string]string{"outcome": "backfill_failed"}})
			}
			if err == nil && progress.Complete {
				return
			}
			timer := time.NewTimer(cfg.BackfillInterval)
			select {
			case <-lifetime.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()
	return func() { cancel(); <-done }
}
