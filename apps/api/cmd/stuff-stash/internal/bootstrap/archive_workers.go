package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
	"time"
)

type archiveDrainer interface {
	Drain(context.Context) (bool, error)
}
type archiveCleaner interface {
	CleanupArchives(context.Context, int) error
}

func startArchiveWorkers(ctx context.Context, worker archiveDrainer, cleaner archiveCleaner, observer ports.Observer, cfg config.ArchiveConfig) func() {
	if !cfg.Enabled {
		return func() {}
	}
	lifetime, cancel := context.WithCancel(ctx)
	var joined sync.WaitGroup
	for range cfg.Concurrency {
		joined.Add(1)
		go func() {
			defer joined.Done()
			for lifetime.Err() == nil {
				work, end := context.WithTimeout(lifetime, cfg.JobTimeout)
				worked, err := worker.Drain(work)
				end()
				if lifetime.Err() != nil {
					return
				}
				if err != nil && observer != nil {
					observer.Record(lifetime, ports.Event{Name: ports.EventArchiveWorkerFailed, Message: "archive worker attempt failed"})
				}
				if worked && err == nil {
					continue
				}
				if !waitArchiveInterval(lifetime, cfg.PollInterval) {
					return
				}
			}
		}()
	}
	joined.Add(1)
	go func() {
		defer joined.Done()
		for lifetime.Err() == nil {
			cleanup, end := context.WithTimeout(lifetime, cfg.CleanupInterval)
			err := cleaner.CleanupArchives(cleanup, 100)
			end()
			if lifetime.Err() != nil {
				return
			}
			if err != nil && observer != nil {
				observer.Record(lifetime, ports.Event{Name: ports.EventArchiveArtifactCleanupFailed, Message: "archive retention cleanup deferred"})
			}
			if !waitArchiveInterval(lifetime, cfg.CleanupInterval) {
				return
			}
		}
	}()
	return func() { cancel(); joined.Wait() }
}
func waitArchiveInterval(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
